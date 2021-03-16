import * as mutations from './mutations.js'

// TODO: use the actual state, instead of making a copy here
let baseState = {
    apiVersion: process.env.VUE_APP_API_VERSION,
    ws: null,
    connected: false,
    errorMessage: "",
    roomId: "test",
    columns: [
        {
            id: "0",
            title: "Good",
            cardStyle: {
                backgroundColor: "bg-danger",
            },
            groups: [{
                id: "default",
                columnId: "0",
                isEditable: false,
                title: "ungrouped cards",
                retroCards: [],
            }],
        },
        {
            id: "1",
            title: "Bad",
            cardStyle: {
                backgroundColor: "bg-primary",
            },
            groups: [{
                id: "default",
                columnId: "1",
                isEditable: false,
                title: "ungrouped cards",
                retroCards: [],
            }],
        },
        {
            id: "3",
            title: "Actions",
            cardStyle: {
                backgroundColor: "bg-success",
            },
            groups: [{
                id: "default",
                columnId: "3",
                isEditable: false,
                title: "ungrouped cards",
                retroCards: [],
            }],
        },
    ],
}

it('adds a new group.', () => {
    let state = JSON.parse(JSON.stringify(baseState))
    let group = {
        id: "test",
        columnId: "0",
        isEditable: true,
        title: "my title",
        retroCards: [],
    };
    mutations.addNewGroup(state, group)

    let expectedState = JSON.parse(JSON.stringify(baseState))
    expectedState.columns[0].groups.push(group)
    expect(expectedState).toStrictEqual(state)
});

it('adds a new card.', () => {
    let state = JSON.parse(JSON.stringify(baseState))
    let card = {
        columnId: "0",
        id: "some-uuid-pk-0",
        message: "my message",
        numVotes: 0,
        isEditable: true,
        groupId: "default",
        isDeleted: false,
        lastModified: 123,
    };

    mutations.addNewRetroCard(state, card)

    let expectedState = JSON.parse(JSON.stringify(baseState))
    expectedState.columns[0].groups[0].retroCards.push(card)
    expect(expectedState).toStrictEqual(state)
});

it('updates group title.', () => {
    let state = JSON.parse(JSON.stringify(baseState))
    let group = {
        id: "test",
        columnId: "0",
        isEditable: true,
        title: "my title",
        retroCards: [],
    };
    state.columns[0].groups.push(group)
    let expectedState = JSON.parse(JSON.stringify(state))

    let updatedTitle = "updated title"
    let groupWithUpdatedTitle = {
        id: group.id,
        columnId: group.columnId,
        isEditable: false,
        title: updatedTitle,
        retroCards: group.retroCards,
    };

    let payload = {
        send: false,
        group: groupWithUpdatedTitle,
    };
    mutations.updateGroupTitle(state, payload)

    expectedState.columns[0].groups[1].title = updatedTitle
    expectedState.columns[0].groups[1].isEditable = false
    expect(expectedState).toStrictEqual(state)
})

it('switches groups.', () => {
    let state = JSON.parse(JSON.stringify(baseState))
    let group = {
        id: "test",
        columnId: "0",
        isEditable: true,
        title: "my title",
        retroCards: [],
    };
    state.columns[0].groups.push(group)

    let anotherGroup = JSON.parse(JSON.stringify(group))
    anotherGroup.title = "another title"
    state.columns[0].groups.push(anotherGroup)

    let expectedState = JSON.parse(JSON.stringify(state))

    let groups = [state.columns[0].groups[1], state.columns[0].groups[2], state.columns[0].groups[0]]
    let payload = {
        columnId: "0",
        groups: groups,
    }
    mutations.switchGroups(state, payload)

    expectedState.columns[0].groups = groups
    expect(expectedState).toStrictEqual(state)
})

it('switches card groups.', () => {
    let state = JSON.parse(JSON.stringify(baseState))
    let group = {
        id: "test",
        columnId: "0",
        isEditable: true,
        title: "my title",
        retroCards: [],
    };
    state.columns[0].groups.push(group)

    let card = {
        columnId: "0",
        id: "some-uuid-pk-0",
        message: "my message",
        numVotes: 0,
        isEditable: false,
        groupId: "default",
        isDeleted: false,
        lastModified: 123,
    }

    state.columns[0].groups[0].retroCards = [card]
    let expectedState = JSON.parse(JSON.stringify(state))

    let movedCard = JSON.parse(JSON.stringify(card))
    let payload = {
        group: state.columns[0].groups[1],
        newRetroCards: [movedCard],
        send: false,
    };

    let expectedCard = JSON.parse(JSON.stringify(movedCard))
    expectedCard.id = "some-uuid-pk-1"
    expectedCard.groupId = "test"

    expectedState.columns[0].groups[0].retroCards[0].isDeleted = true
    expectedState.columns[0].groups[1].retroCards.push(expectedCard)

    mutations.switchCardGroup(state, payload)

    expect(state.columns[0].groups[0].retroCards[0].lastModified).not.toEqual(123)
    expect(state.columns[0].groups[1].retroCards[0].lastModified).not.toEqual(123)
    state.columns[0].groups[0].retroCards[0].lastModified = 123
    state.columns[0].groups[1].retroCards[0].lastModified = 123
    expect(expectedState).toStrictEqual(state)
})

it('switches card in same group.', () => {
    let state = JSON.parse(JSON.stringify(baseState))
    let card = {
        columnId: "0",
        id: "some-uuid-pk-0",
        message: "my message",
        numVotes: 0,
        isEditable: false,
        groupId: "default",
        isDeleted: false,
        lastModified: 123,
    };

    let anotherCard = JSON.parse(JSON.stringify(card))
    anotherCard.id = "another-uuid-pk-0"

    state.columns[0].groups[0].retroCards = [card, anotherCard]

    let cardCopy = JSON.parse(JSON.stringify(card))
    let anotherCardCopy = JSON.parse(JSON.stringify(anotherCard))

    let payload = {
        group: state.columns[0].groups[0],
        newRetroCards: [anotherCardCopy, cardCopy]
    }

    mutations.switchCardInSameGroup(state, payload)

    let expectedState = JSON.parse(JSON.stringify(baseState))
    expectedState.columns[0].groups[0].retroCards = [anotherCardCopy, cardCopy]
    expect(expectedState).toStrictEqual(state)
});

it('it updates retro card message.', () => {
    let state = JSON.parse(JSON.stringify(baseState))
    let card = {
        columnId: "0",
        id: "some-uuid-pk-0",
        message: "my message",
        numVotes: 0,
        isEditable: true,
        groupId: "default",
        isDeleted: false,
        lastModified: 123,
    };

    state.columns[0].groups[0].retroCards = [card]
    let cardCopy = JSON.parse(JSON.stringify(card))
    cardCopy.isEditable = false
    cardCopy.message = "updated message"

    let payload = {
        card: cardCopy,
        action: null,
        send: false,
    }

    mutations.updateRetroCard(state, payload)

    let expectedState = JSON.parse(JSON.stringify(baseState))
    expectedState.columns[0].groups[0].retroCards[0] = cardCopy
    expect(expectedState).toStrictEqual(state)
});

it('it updates retro card votes.', () => {
    let state = JSON.parse(JSON.stringify(baseState))
    let originalCard = {
        columnId: "0",
        id: "some-uuid-pk-0",
        message: "my message",
        numVotes: 0,
        isEditable: false,
        groupId: "default",
        isDeleted: true,
        lastModified: 123,
    }
    let cardToUpvote = JSON.parse(JSON.stringify(originalCard))
    cardToUpvote.id = "some-uuid-pk-1"
    cardToUpvote.isDeleted = false
    state.columns[0].groups[0].retroCards = [originalCard, cardToUpvote]

    let expectedState = JSON.parse(JSON.stringify(state))

    let cardToUpvoteCopy = JSON.parse(JSON.stringify(cardToUpvote))
    cardToUpvoteCopy.numVotes++

    let payload = {
        card: cardToUpvoteCopy,
        action: {
            title: "upVote",
            oldCard: cardToUpvote,
            newCard: cardToUpvoteCopy,
        },
        send: false,
    }

    mutations.updateRetroCard(state, payload)

    for (let i = 0; i < expectedState.columns[0].groups[0].retroCards.length; i++) {
        expectedState.columns[0].groups[0].retroCards[i].numVotes++
    }

    expect(expectedState).toStrictEqual(state)
});

it('it updates retro card votes to the max in the chain.', () => {
    let state = JSON.parse(JSON.stringify(baseState))
    let originalCard = {
        columnId: "0",
        id: "some-uuid-pk-0",
        message: "my message",
        numVotes: 5,
        isEditable: false,
        groupId: "default",
        isDeleted: true,
        lastModified: 123,
    }
    let cardToUpvote = JSON.parse(JSON.stringify(originalCard))
    cardToUpvote.id = "some-uuid-pk-1"
    cardToUpvote.isDeleted = false
    cardToUpvote.numVotes = 1
    state.columns[0].groups[0].retroCards = [originalCard, cardToUpvote]

    let expectedState = JSON.parse(JSON.stringify(state))

    let cardToUpvoteCopy = JSON.parse(JSON.stringify(cardToUpvote))
    cardToUpvoteCopy.numVotes++

    let payload = {
        card: cardToUpvoteCopy,
        action: {
            title: "upVote",
            oldCard: cardToUpvote,
            newCard: cardToUpvoteCopy,
        },
        send: false,
    }

    mutations.updateRetroCard(state, payload)

    for (let i = 0; i < expectedState.columns[0].groups[0].retroCards.length; i++) {
        expectedState.columns[0].groups[0].retroCards[i].numVotes = 5
    }

    expect(expectedState).toStrictEqual(state)
});

// TODO: test updateColumns
// updateColumns
// new cards are appended at the bottom
// cards with the same id are updated
// duplicates are handled by last modified & by deleted
// max upvotes in the chain is maintained