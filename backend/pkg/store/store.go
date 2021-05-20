package store

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/safe-waters/retro-simply/backend/pkg/client"
	"github.com/safe-waters/retro-simply/backend/pkg/data"
	"go.opentelemetry.io/otel"
)

var tr = otel.Tracer("pkg/store")

const (
	pPrefix = "password"
	sPrefix = "state"
)

type DatabaseGetWatchSetter interface {
	Get(ctx context.Context, key string) client.StrResult
	Watch(ctx context.Context, fn func(*redis.Tx) error, keys ...string) error
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) client.BoolResult
}

type S struct{ d DatabaseGetWatchSetter }

func New(d DatabaseGetWatchSetter) *S { return &S{d: d} }

func (s *S) State(ctx context.Context, rId string) (*data.State, error) {
	ctx, span := tr.Start(ctx, "get state")
	defer span.End()

	k := s.getKey(sPrefix, rId)

	v, err := s.d.Get(ctx, k).Result()
	if err != nil {
		span.RecordError(err)

		switch err {
		case redis.Nil:
			err := DataDoesNotExistError{err}
			span.RecordError(err)

			return nil, err
		default:
			return nil, err
		}
	}

	var st data.State

	b := []byte(v)
	if err := json.Unmarshal(b, &st); err != nil {
		span.RecordError(err)
		return nil, err
	}

	return &st, nil
}

type retroCardWithIndex struct {
	retroCard *data.RetroCard
	index     int
}

func (s *S) mergeState(ctx context.Context, os *data.State, st *data.State) (*data.State, error) {
	_, span := tr.Start(ctx, "merge state")
	defer span.End()

	var ms data.State

	// copy oldState into mergedState, so mergedState can be used as a
	// starting point and changed without changing oldState
	msByt, err := json.Marshal(os)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	if err := json.Unmarshal(msByt, &ms); err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Adding new columns is not allowed
	if len(os.Columns) != len(st.Columns) {
		err := fmt.Errorf(
			"expected old state columns %d, got state columns %d",
			len(os.Columns),
			len(st.Columns),
		)
		span.RecordError(err)

		return nil, err
	}

	// Save any new cards, which will be used to make sure upvotes
	// are correct, later
	newCardIds := map[string]struct{}{}

	for i := 0; i < len(st.Columns); i++ {
		// changing the order of columns is not allowed
		if os.Columns[i].Id != st.Columns[i].Id {
			err := fmt.Errorf(
				"expected old state columns id %s, got state columns id %s",
				os.Columns[i].Id,
				st.Columns[i].Id,
			)
			span.RecordError(err)

			return nil, err
		}

		for j := 0; j < len(st.Columns[i].Groups); j++ {
			// check if the oldState has the group in state. If it
			// does not, add it to the merged state
			var gFound bool

			// merge all the groups that exist in oldState and state
			for k := 0; k < len(os.Columns[i].Groups); k++ {
				if st.Columns[i].Groups[j].Id == os.Columns[i].Groups[k].Id {
					gFound = true

					for l := 0; l < len(st.Columns[i].Groups[j].RetroCards); l++ {
						// Check if oldState has the retro card. If it
						// does not, add it to the merged state.
						//
						// Cards are never deleted, instead they have an "isDeleted"
						// field. Therefore, we only have to care about
						// additions, as there never will be "hard" deletes.
						//
						// If one state has a card and another does not,
						// they are both valid, but represent states at
						// different times (for instance, n and n+1).
						var rFound bool

						for a := 0; a < len(os.Columns[i].Groups[k].RetroCards); a++ {
							if st.Columns[i].Groups[j].RetroCards[l].Id == os.Columns[i].Groups[k].RetroCards[a].Id {
								rFound = true

								// set the mergedState card to state's card
								ms.Columns[i].Groups[k].RetroCards[a] = st.Columns[i].Groups[j].RetroCards[l]

								// take the max of the oldState and state's numVotes,
								// as numVotes can never decrease.
								if os.Columns[i].Groups[k].RetroCards[a].NumVotes > ms.Columns[i].Groups[k].RetroCards[a].NumVotes {
									ms.Columns[i].Groups[k].RetroCards[a].NumVotes = os.Columns[i].Groups[k].RetroCards[a].NumVotes
								}

								// if the oldState's card was deleted, make sure it is still deleted
								if os.Columns[i].Groups[k].RetroCards[a].IsDeleted {
									ms.Columns[i].Groups[k].RetroCards[a].IsDeleted = true
								}
							}
						}

						if !rFound {
							// add the new card to the group
							ms.Columns[i].Groups[k].RetroCards = append(
								ms.Columns[i].Groups[k].RetroCards,
								st.Columns[i].Groups[j].RetroCards[l],
							)

							// save the new card
							newCardIds[st.Columns[i].Groups[j].RetroCards[l].Id] = struct{}{}
						}
					}
				}
			}

			if !gFound {
				// add the new group to the mergedState
				ms.Columns[i].Groups = append(
					ms.Columns[i].Groups,
					st.Columns[i].Groups[j],
				)

				// save the new cards
				for a := 0; a < len(st.Columns[i].Groups[j].RetroCards); a++ {
					newCardIds[st.Columns[i].Groups[j].RetroCards[a].Id] = struct{}{}
				}
			}
		}
	}

	// After merging groups, there may be multiple cards with the same ID
	// in different groups of mergedState.

	// store all of the potential duplicates
	dups := map[string][]*retroCardWithIndex{}

	for i := 0; i < len(ms.Columns); i++ {
		for j := 0; j < len(ms.Columns[i].Groups); j++ {
			for k := 0; k < len(ms.Columns[i].Groups[j].RetroCards); k++ {
				dups[ms.Columns[i].Groups[j].RetroCards[k].Id] = append(
					dups[ms.Columns[i].Groups[j].RetroCards[k].Id],
					&retroCardWithIndex{retroCard: ms.Columns[i].Groups[j].RetroCards[k], index: k},
				)
			}
		}
	}

	// Store all of the duplicates that should be deleted from mergedState,
	// along with their retroCard indexes, so they can be deleted from
	// mergedState in descending order so as not to change the position
	// of the cards in mergedState.
	var cardsToDelete []*retroCardWithIndex

	for _, cards := range dups {
		// if there is more than one card, there is a duplicate that needs
		// to be deleted
		if len(cards) > 1 {
			var cardToKeepIndex int

			for i := 1; i < len(cards); i++ {
				// When deleting a card, always make sure the card to keep has the max numvotes
				if cards[cardToKeepIndex].retroCard.NumVotes < cards[i].retroCard.NumVotes {
					cards[cardToKeepIndex].retroCard.NumVotes = cards[i].retroCard.NumVotes
				} else {
					cards[i].retroCard.NumVotes = cards[cardToKeepIndex].retroCard.NumVotes
				}

				switch {
				// if both cards are deleted, keep the last modified card
				case cards[cardToKeepIndex].retroCard.IsDeleted && cards[i].retroCard.IsDeleted:
					if cards[cardToKeepIndex].retroCard.LastModified < cards[i].retroCard.LastModified {
						cardToKeepIndex = i
					}
				// if card to keep is not deleted, but the other one is deleted, keep the one
				// that is deleted
				case !cards[cardToKeepIndex].retroCard.IsDeleted && cards[i].retroCard.IsDeleted:
					cardToKeepIndex = i
				// if both cards are deleted, keep the last modified card
				case !cards[cardToKeepIndex].retroCard.IsDeleted && !cards[i].retroCard.IsDeleted:
					if cards[cardToKeepIndex].retroCard.LastModified < cards[i].retroCard.LastModified {
						cardToKeepIndex = i
					}
				}
			}

			// Remove card to keep from the cards slice, so only cards
			// that should be deleted from mergedState will actually be
			// deleted.
			cards[len(cards)-1], cards[cardToKeepIndex] = cards[cardToKeepIndex], cards[len(cards)-1]
			cardsToDelete = append(cardsToDelete, cards[:len(cards)-1]...)
		}
	}

	// If there are cards to delete, sort them by group id and by descending
	// order of their retroCard indexes, so that deleting from the front
	// of the slice does not change the position of a future retro card to
	// delete in the slice.
	if len(cardsToDelete) > 0 {
		sort.Slice(cardsToDelete, func(i, j int) bool {
			if cardsToDelete[i].retroCard.GroupId > cardsToDelete[j].retroCard.GroupId {
				return false
			} else if cardsToDelete[i].retroCard.GroupId < cardsToDelete[j].retroCard.GroupId {
				return true
			} else {
				if cardsToDelete[i].index < cardsToDelete[j].index {
					return false
				} else {
					return true
				}
			}
		})

		// delete duplicates from mergedState
		for i := 0; i < len(ms.Columns); i++ {
			for j := 0; j < len(ms.Columns[i].Groups); j++ {
				for k := 0; k < len(cardsToDelete); k++ {
					// checking more than just the group ID matters because currently,
					// every column has a group with the ID of default.
					// TODO: in the future change this so all groups have
					// a unique id.
					if cardsToDelete[k].retroCard.GroupId == ms.Columns[i].Groups[j].Id &&
						len(ms.Columns[i].Groups[j].RetroCards) > cardsToDelete[k].index &&
						ms.Columns[i].Groups[j].RetroCards[cardsToDelete[k].index].Id == cardsToDelete[k].retroCard.Id {

						ms.Columns[i].Groups[j].RetroCards = append(
							ms.Columns[i].Groups[j].RetroCards[:cardsToDelete[k].index],
							ms.Columns[i].Groups[j].RetroCards[cardsToDelete[k].index+1:]...,
						)
					}
				}
			}
		}
	}

	// Store all of the cards, so they can be easily changed if you
	// need to traverse the previous / next cards.
	//
	// Card IDs are of the form "{uuid}-pk-{number}". Moving a card from
	// one group to another sets the "isDeleted" property to true on the
	// card in the original group and creates the "next" card in the new
	// group by incrementing the number after pk. A card "chain" would be
	// all of the cards that share the same uuid.
	cardsById := map[string]*data.RetroCard{}

	for i := 0; i < len(ms.Columns); i++ {
		for j := 0; j < len(ms.Columns[i].Groups); j++ {
			for k := 0; k < len(ms.Columns[i].Groups[j].RetroCards); k++ {
				cardsById[ms.Columns[i].Groups[j].RetroCards[k].Id] = ms.Columns[i].Groups[j].RetroCards[k]
			}
		}
	}

	// Make sure all new cards update their card chains.
	// We want to make sure that all the cards
	// in the chain have the same number of upvotes, which will always
	// be the maximum number of any member in the chain.
	//
	// Maintaining an equal number of upvotes makes logical sense, since
	// all the cards are really the same. It also provides peace of mind
	// that a card that isDeleted could not possibly have more upvotes
	// than its successor.
	for id := range newCardIds {
		numVotes := s.getMaxNumUpvotesInCardChain(id, cardsById)
		s.applyNumUpvotesToCardChain(id, cardsById, numVotes)
	}

	if st.Action != nil {
		switch st.Action.Title {
		case "upVote":
			numVotes := s.getMaxNumUpvotesInCardChain(st.Action.NewCard.Id, cardsById)

			for i := 0; i < len(os.Columns); i++ {
				for j := 0; j < len(os.Columns[i].Groups); j++ {
					for k := 0; k < len(os.Columns[i].Groups[j].RetroCards); k++ {
						if os.Columns[i].Groups[j].RetroCards[k].Id == st.Action.NewCard.Id {
							// If oldState's numVotes is ahead of action's old card numvotes,
							// add 1 to avoid a lost update.
							if os.Columns[i].Groups[j].RetroCards[k].NumVotes > st.Action.OldCard.NumVotes {
								numVotes++

								break
							}
						}
					}
				}
			}

			// Ensure the numVotes is reflected in every card in the chain.
			s.applyNumUpvotesToCardChain(st.Action.NewCard.Id, cardsById, numVotes)
		}
	}

	return &ms, nil
}

func (s *S) StoreState(ctx context.Context, st *data.State) (*data.State, error) {
	ctx, span := tr.Start(ctx, "store state")
	defer span.End()

	var ms *data.State
	k := s.getKey(sPrefix, st.RoomId)

	txf := func(tx *redis.Tx) error {
		ctx, span := tr.Start(ctx, "transaction")
		defer span.End()

		var err error

		// If oldState does not exist, use state. Otherwise,
		// merge oldState and state.
		os, err := s.State(ctx, st.RoomId)
		if err != nil {
			span.RecordError(err)

			switch err.(type) {
			case DataDoesNotExistError:
				ms = st
			default:
				return err
			}
		}

		if ms == nil {
			ms, err = s.mergeState(ctx, os, st)
			if err != nil {
				span.RecordError(err)
				return err
			}
		}

		msByt, err := json.Marshal(ms)
		if err != nil {
			span.RecordError(err)
			return err
		}

		// Store the mergedState, returning a redis.TxFailedErr if the
		// value stored at the key has changed.
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Set(ctx, k, msByt, 0)
			return nil
		})

		if err != nil {
			span.RecordError(err)
		}

		return err
	}

	// Optimistic locking - try to store mergedState up to maxRetries.
	//
	// The transaction will fail if the value stored at the key
	// changes before the transaction stores mergedState in Redis.
	//
	// Example:
	// https://pkg.go.dev/github.com/go-redis/redis/v8#Client.Watch
	//
	// More information about watch:
	// https://redislabs.com/blog/you-dont-need-transaction-rollbacks-in-redis/
	var err error
	const retries = 10000

	for i := 0; i < retries; i++ {
		err = s.d.Watch(ctx, txf, k)
		if err != nil {
			span.RecordError(err)

			switch err {
			case redis.TxFailedErr:
				// If the transaction failed, try again
				continue
			default:
				// If failed for any reason unrelated to optimistic locking,
				// return err
				return nil, err
			}
		}

		return ms, nil
	}

	return nil, err
}

const pkPrefix = "-pk-"

func (s *S) getMaxNumUpvotesInCardChain(id string, cardsById map[string]*data.RetroCard) uint {
	idWithoutPk := s.getIdWithoutPk(id)
	numVotes := cardsById[id].NumVotes
	pk := 0

	for {
		currId := fmt.Sprintf("%s%s%d", idWithoutPk, pkPrefix, pk)

		card, ok := cardsById[currId]
		if !ok {
			break
		}

		if card.NumVotes > numVotes {
			numVotes = card.NumVotes
		}

		pk++
	}

	return numVotes
}

func (s *S) applyNumUpvotesToCardChain(id string, cardsById map[string]*data.RetroCard, numVotes uint) {
	idWithoutPk := s.getIdWithoutPk(id)
	pk := 0

	for {
		currId := fmt.Sprintf("%s%s%d", idWithoutPk, pkPrefix, pk)

		card, ok := cardsById[currId]
		if !ok {
			break
		}

		card.NumVotes = numVotes
		pk++
	}
}

func (s *S) getIdWithoutPk(id string) string {
	return id[:strings.LastIndex(id, pkPrefix)]
}

func (s *S) StoreHashedPassword(ctx context.Context, rId, h string) error {
	ctx, span := tr.Start(ctx, "store hashed password")
	defer span.End()

	k := s.getKey(pPrefix, rId)

	didSet, err := s.d.SetNX(ctx, k, []byte(h), 0).Result()
	if err != nil {
		span.RecordError(err)
		return err
	}

	if !didSet {
		err := DataAlreadyExistsError{fmt.Errorf("room '%s' already exists", rId)}
		span.RecordError(err)

		return err
	}

	return nil
}

func (s *S) HashedPassword(ctx context.Context, rId string) (string, error) {
	ctx, span := tr.Start(ctx, "get hashed password")
	defer span.End()

	k := s.getKey(pPrefix, rId)

	h, err := s.d.Get(ctx, k).Result()
	if err != nil {
		span.RecordError(err)

		switch err {
		case redis.Nil:
			err := DataDoesNotExistError{fmt.Errorf("room '%s' does not exist", rId)}
			span.RecordError(err)

			return "", err
		default:
			return "", err
		}
	}

	return h, nil
}

func (s *S) getKey(prefix string, identifier string) string {
	return fmt.Sprintf("%s%s", prefix, identifier)
}
