import * as helpers from './mutationsHelpers.js'

export function connect(state) {
  state.ws = new WebSocket("wss://" + window.location.host + "/api/" + state.apiVersion + "/retrospectives/" + state.roomId);
}

export function setConnected(state, status) {
  state.connected = status
}

export function setErrorMessage(state, message) {
  state.errorMessage = message
}

export function addNewGroup(state, group) {
  let newState = JSON.parse(JSON.stringify(state))
  for (let i = 0; i < newState.columns.length; i++) {
    if (newState.columns[i].id === group.columnId) {
      let defaultIndex = 0
      for (let j = 0; j < newState.columns[i].groups.length; j++) {
        if (newState.columns[i].groups[j].id === "default") {
          defaultIndex = j
        }
      }
      newState.columns[i].groups.splice(defaultIndex + 1, 0, group)
      helpers.updateLocalColumns(state, newState.columns)
      return
    }
  }
}

export function addNewRetroCard(state, card) {
  let newState = JSON.parse(JSON.stringify(state))
  for (let i = 0; i < newState.columns.length; i++) {
    if (newState.columns[i].id === card.columnId) {
      for (let j = 0; j < newState.columns[i].groups.length; j++) {
        if (newState.columns[i].groups[j].id === card.groupId) {
          newState.columns[i].groups[j].retroCards.unshift(card)
          helpers.updateLocalColumns(state, newState.columns)
          return
        }
      }
    }
  }
}

export function updateGroupTitle(state, payload) {
  let { send, group } = payload

  let newState = JSON.parse(JSON.stringify(state))
  for (let i = 0; i < newState.columns.length; i++) {
    if (newState.columns[i].id === group.columnId) {
      for (let j = 0; j < newState.columns[i].groups.length; j++) {
        if (newState.columns[i].groups[j].id === group.id) {
          newState.columns[i].groups[j] = group
          helpers.updateLocalColumns(state, newState.columns)
          if (send) {
            helpers.sendState(state.ws, newState)
          }
          return
        }
      }
    }
  }
}

export function switchGroups(state, payload) {
  let newState = JSON.parse(JSON.stringify(state))
  let { columnId, groups } = payload
  for (let i = 0; i < newState.columns.length; i++) {
    if (newState.columns[i].id === columnId) {
      newState.columns[i].groups = groups
      break
    }
  }

  helpers.updateLocalColumns(state, newState.columns)
}

export function switchCardGroup(state, payload) {
  let { group, newRetroCards, send } = payload
  let newCard = JSON.parse(JSON.stringify(newRetroCards[0]))

  let changeOccurred = false
  for (let i = 0; i < newRetroCards.length; i++) {
    let cardFound = false
    for (let j = 0; j < group.retroCards.length; j++) {
      if (newRetroCards[i].id === group.retroCards[j].id) {
        cardFound = true
      }
    }

    if (!cardFound) {
      // save the original new card, so that we have the group id
      newCard = JSON.parse(JSON.stringify(newRetroCards[i]))

      // set the group id for the new card to the group it will become part of
      newRetroCards[i].groupId = group.id

      // set last modified time
      newRetroCards[i].lastModified = Date.now()

      // give the new card a new id
      let pkIndex = newCard.id.lastIndexOf("-pk-");
      let idNum = parseInt(newCard.id.substr(pkIndex + "-pk-".length)) + 1;
      newRetroCards[i].id = newCard.id.substr(0, pkIndex + "-pk-".length) + idNum.toString();

      // if the new id matches any other retroCard, do nothing
      for (let k = 0; k < state.columns.length; k++) {
        for (let l = 0; l < state.columns[k].groups.length; l++) {
          for (let a = 0; a < state.columns[k].groups[l].retroCards.length; a++) {
            if (state.columns[k].groups[l].retroCards[a].id === newRetroCards[i].id) {
              return
            }
          }
        }
      }

      changeOccurred = true
      break
    }
  }

  if (changeOccurred) {
    let newState = JSON.parse(JSON.stringify(state))
    for (let i = 0; i < newState.columns.length; i++) {
      if (newState.columns[i].id === newCard.columnId) {
        let hasDeletedCard = false
        let hasReplacedCards = false
        for (let j = 0; j < newState.columns[i].groups.length; j++) {
          if (!hasDeletedCard && newState.columns[i].groups[j].id === newCard.groupId) {
            for (let k = 0; k < newState.columns[i].groups[j].retroCards.length; k++) {
              if (newState.columns[i].groups[j].retroCards[k].id === newCard.id) {
                // find the card that would be deleted from the old group
                // and set a property for deleted
                newState.columns[i].groups[j].retroCards[k].isDeleted = true
                newState.columns[i].groups[j].retroCards[k].lastModified = Date.now()
                hasDeletedCard = true
                break
              }
            }
          }

          if (!hasReplacedCards && newState.columns[i].groups[j].id === group.id) {
            // set the group to the new retro cards
            newState.columns[i].groups[j].retroCards = newRetroCards
            hasReplacedCards = true
          }
        }
      }
    }

    // update the local state so that order is maintained locally
    helpers.updateLocalColumns(state, newState.columns)

    // since this modifies cards across groups, send the state to others
    if (send) {
      helpers.sendState(state.ws, newState)
    }
  }
}

export function switchCardInSameGroup(state, payload) {
  let { group, newRetroCards } = payload
  let newState = JSON.parse(JSON.stringify(state))

  for (let i = 0; i < newState.columns.length; i++) {
    if (newState.columns[i].id === group.columnId) {
      for (let j = 0; j < newState.columns[i].groups.length; j++) {
        if (newState.columns[i].groups[j].id === group.id) {
          newState.columns[i].groups[j].retroCards = newRetroCards
          break
        }
      }
    }
  }

  helpers.updateLocalColumns(state, newState.columns)
}

export function updateRetroCard(state, payload) {
  let { card, action, send } = payload

  let newState = JSON.parse(JSON.stringify(state))
  newState.action = action

  loop:
  for (let i = 0; i < newState.columns.length; i++) {
    if (newState.columns[i].id === card.columnId) {
      for (let j = 0; j < newState.columns[i].groups.length; j++) {
        if (newState.columns[i].groups[j].id === card.groupId) {
          for (let k = 0; k < newState.columns[i].groups[j].retroCards.length; k++) {
            if (newState.columns[i].groups[j].retroCards[k].id === card.id) {
              newState.columns[i].groups[j].retroCards[k] = card
              break loop
            }
          }
        }
      }
    }
  }

  if (newState.action && newState.action.title === "upVote") {
    let cardsById = {}
    for (let i = 0; i < newState.columns.length; i++) {
      for (let j = 0; j < newState.columns[i].groups.length; j++) {
        for (let k = 0; k < newState.columns[i].groups[j].retroCards.length; k++) {
          cardsById[newState.columns[i].groups[j].retroCards[k].id] = newState.columns[i].groups[j].retroCards[k]
        }
      }
    }

    let numVotes = getMaxNumUpvotesInCardChain(action.newCard.id, cardsById)
    applyNumUpvotesToCardChain(action.newCard.id, cardsById, numVotes)
  }

  helpers.updateLocalColumns(state, newState.columns)
  if (send) {
    helpers.sendState(state.ws, newState)
  }
  return
}

export function updateColumns(state, columns) {
  let newNonEditableCards = []
  for (let i = 0; i < columns.length; i++) {
    let { mergedGroups, newCards } = helpers.mergeGroups(state.columns[i].groups, columns[i].groups)
    columns[i].groups = mergedGroups
    for (let j = 0; j < newCards.length; j++) {
      newNonEditableCards.push(newCards[j])
    }
  }

  let duplicates = {}
  for (let i = 0; i < columns.length; i++) {
    for (let j = 0; j < columns[i].groups.length; j++) {
      for (let k = 0; k < columns[i].groups[j].retroCards.length; k++) {
        let cardWithIndex = JSON.parse(JSON.stringify(columns[i].groups[j].retroCards[k]))
        cardWithIndex.index = k

        if (!(columns[i].groups[j].retroCards[k].id in duplicates)) {
          duplicates[cardWithIndex.id] = [cardWithIndex]
        } else {
          duplicates[cardWithIndex.id].push(cardWithIndex)
        }
      }
    }
  }

  let cardsToDelete = []
  for (const key of Object.keys(duplicates)) {
    if (duplicates[key].length > 1) {
      let cardToKeepIndex = 0

      for (let i = 1; i < duplicates[key].length; i++) {
        if (duplicates[key][cardToKeepIndex].numVotes < duplicates[key][i].numVotes) {
          duplicates[key][cardToKeepIndex].numVotes = duplicates[key][i].numVotes
        } else {
          duplicates[key][i].numVotes = duplicates[key][cardToKeepIndex].numVotes
        }

        if (duplicates[key][cardToKeepIndex].isDeleted && duplicates[key][i].isDeleted) {
          if (duplicates[key][cardToKeepIndex].lastModified < duplicates[key][i].lastModified) {
            cardToKeepIndex = i
          }

          continue
        }

        if (!duplicates[key][cardToKeepIndex].isDeleted && duplicates[key][i].isDeleted) {
          cardToKeepIndex = i
          continue
        }

        if (!duplicates[key][cardToKeepIndex].isDeleted && !duplicates[key][i].isDeleted) {
          if (duplicates[key][cardToKeepIndex].lastModified < duplicates[key][i].lastModified) {
            cardToKeepIndex = i
          }

          continue
        }
      }

      cardsToDelete = [...cardsToDelete, ...duplicates[key].splice(cardToKeepIndex, 1)]
    }
  }

  // sort cards by groupId and then last index first
  if (cardsToDelete.length > 0) {
    cardsToDelete.sort(function (a, b) {
      if (a.groupId > b.groupId) {
        return 1
      } else if (a.groupId < b.groupId) {
        return -1
      } else {
        if (a.index < b.index) {
          return 1
        } else {
          return -1
        }
      }
    })

    for (let i = 0; i < columns.length; i++) {
      for (let j = 0; j < columns[i].groups.length; j++) {
        for (let k = 0; k < cardsToDelete.length; k++) {
          if (cardsToDelete[k].groupId === columns[i].groups[j].id &&
            columns[i].groups[j].retroCards.length > cardsToDelete[k].index &&
            columns[i].groups[j].retroCards[cardsToDelete[k].index].id == cardsToDelete[k].id) {
            columns[i].groups[j].retroCards.splice(cardsToDelete[k].index, 1)
          }
        }
      }
    }
  }

  // get all the cards into a dictionary
  let cardsById = {}
  for (let i = 0; i < columns.length; i++) {
    for (let j = 0; j < columns[i].groups.length; j++) {
      for (let k = 0; k < columns[i].groups[j].retroCards.length; k++) {
        cardsById[columns[i].groups[j].retroCards[k].id] = columns[i].groups[j].retroCards[k]
      }
    }
  }

  // find the max of everything in the chain, and use that
  for (let i = 0; i < newNonEditableCards.length; i++) {
    let numVotes = getMaxNumUpvotesInCardChain(newNonEditableCards[i].id, cardsById)
    applyNumUpvotesToCardChain(newNonEditableCards[i].id, cardsById, numVotes)
  }

  state.columns = columns
}

const pkPrefix = "-pk-"

function getMaxNumUpvotesInCardChain(id, cardsById) {
  let idWithoutPk = id.substring(0, id.lastIndexOf(pkPrefix))

  let numVotes = cardsById[id].numVotes
  let pk = 0

  for (; ;) {
    let currId = idWithoutPk + pkPrefix + pk.toString()
    if (!(currId in cardsById)) {
      break
    }

    let card = cardsById[currId]

    if (card.numVotes > numVotes) {
      numVotes = card.numVotes
    }

    pk++
  }

  return numVotes
}


function applyNumUpvotesToCardChain(id, cardsById, numVotes) {
  let idWithoutPk = id.substring(0, id.lastIndexOf(pkPrefix))
  let pk = 0

  for (; ;) {
    let currId = idWithoutPk + pkPrefix + pk.toString()
    if (!(currId in cardsById)) {
      break
    }

    let card = cardsById[currId]
    card.numVotes = numVotes
    pk++
  }
}

export function sortByNumVotes(state) {
  let newState = JSON.parse(JSON.stringify(state))
  for (let i = 0; i < newState.columns.length; i++) {
    newState.columns[i].groups.sort(function (a, b) {
      if (b.id === "default") {
        return 1
      }

      if (a.id == "default") {
        return -1
      }

      let aVotes = 0;
      for (let i = 0; i < a.retroCards.length; i++) {
        if (!a.retroCards[i].isDeleted) {
          aVotes += a.retroCards[i].numVotes;
        }
      }

      let bVotes = 0;
      for (let i = 0; i < b.retroCards.length; i++) {
        if (!b.retroCards[i].isDeleted) {
          bVotes += b.retroCards[i].numVotes;
        }
      }

      return bVotes - aVotes
    })

    for (let j = 0; j < newState.columns[i].groups.length; j++) {
      newState.columns[i].groups[j].retroCards.sort(function (a, b) {
        if (b.isEditable) {
          return 1
        }

        if (a.isEditable) {
          return -1
        }

        return b.numVotes - a.numVotes
      })
    }
  }

  helpers.updateLocalColumns(state, newState.columns)
  return
}