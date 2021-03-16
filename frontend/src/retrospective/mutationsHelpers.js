export function mergeGroups(groups, newGroups) {
  let newNonEditableRetroCards = []

  for (let i = 0; i < newGroups.length; i++) {
    let groupFound = false
    for (let j = 0; j < groups.length; j++) {
      if (newGroups[i].id === groups[j].id) {
        groupFound = true
      }
    }

    if (!groupFound) {
      for (let j = 0; j < newGroups[i].retroCards.length; j++) {
        if (!newGroups[i].retroCards[j].isEditable) {
          newNonEditableRetroCards.push(newGroups[i].retroCards[j])
        }
      }
    }
  }

  let concatenatedGroups = [...groups, ...newGroups]
  let set = new Set()
  let mergedGroups = []

  // keep local groups first before new groups
  for (let i = 0; i < concatenatedGroups.length; i++) {
    if (!set.has(concatenatedGroups[i].id)) {
      set.add(concatenatedGroups[i].id)
      mergedGroups.push(concatenatedGroups[i])
    }
  }

  // update the groups to use the values from the new groups, if they exist
  for (let i = 0; i < mergedGroups.length; i++) {
    for (let j = 0; j < newGroups.length; j++) {
      if (mergedGroups[i].id === newGroups[j].id) {
        let {mergedRetroCards, newCards} = mergeRetroCards(mergedGroups[i].retroCards, newGroups[j].retroCards)

        for (let k = 0; k < newCards.length; k++) {
          newNonEditableRetroCards.push(newCards[k])
        }

        let group = {
          id: newGroups[j].id,
          columnId: newGroups[j].columnId,
          title: newGroups[j].title,
          retroCards: mergedRetroCards,
        }
        mergedGroups[i] = group
      }
    }
  }

  return {
    "mergedGroups": mergedGroups,
    "newCards": newNonEditableRetroCards,
  }
}

export function mergeRetroCards(retroCards, newRetroCards) {
  let editableRetroCardsNotInNewRetroCards = []
  for (let i = 0; i < retroCards.length; i++) {
    if (retroCards[i].isEditable) {
      let cardFound = false

      for (let j = 0; j < newRetroCards.length; j++) {
        if (retroCards[i].id === newRetroCards[j].id) {
          cardFound = true
          break
        }
      }

      if (!cardFound) {
        editableRetroCardsNotInNewRetroCards.push(retroCards[i])
      }
    }
  }

  // keep new editable also on top
  let editableNewRetroCards = newRetroCards.filter(card => card.isEditable)

  // update the cards to use the values from the new cards
  let nonEditableUpdatedRetroCards = []
  for (let i = 0; i < retroCards.length; i++) {
    let cardFound = false
    for (let j = 0; j < newRetroCards.length; j++) {
      if (!newRetroCards[j].isEditable && retroCards[i].id === newRetroCards[j].id) {
        cardFound = true
        newRetroCards[j].numVotes = Math.max(newRetroCards[j].numVotes, retroCards[i].numVotes)
        nonEditableUpdatedRetroCards.push(newRetroCards[j])
      }
    }

    if (!cardFound && !retroCards[i].isEditable) {
      nonEditableUpdatedRetroCards.push(retroCards[i])
    }
  }

  // add new cards to the bottom
  let nonEditableNewRetroCardsNotInRetroCards = []
  for (let i = 0; i < newRetroCards.length; i++) {
    if (!newRetroCards[i].isEditable) {
      let cardFound = false

      for (let j = 0; j < retroCards.length; j++) {
        if (newRetroCards[i].id === retroCards[j].id) {
          cardFound = true
          break
        }
      }
      if (!cardFound) {
        nonEditableNewRetroCardsNotInRetroCards.push(newRetroCards[i])
      }
    }
  }

  let mergedRetroCards = [
    ...editableRetroCardsNotInNewRetroCards,
    ...editableNewRetroCards,
    ...nonEditableUpdatedRetroCards,
    ...nonEditableNewRetroCardsNotInRetroCards
  ]

  return {
    "mergedRetroCards": mergedRetroCards,
    "newCards": nonEditableNewRetroCardsNotInRetroCards,
  }
}

export function updateLocalColumns(state, columns) {
  state.columns = columns
}

export function sendState(ws, state) {
  let newState = JSON.parse(JSON.stringify(state))
  for (let i = 0; i < newState.columns.length; i++) {
    newState.columns[i].groups = newState.columns[i].groups.filter(group => !group.isEditable)
    for (let j = 0; j < newState.columns[i].groups.length; j++) {
      newState.columns[i].groups[j].retroCards = newState.columns[i].groups[j].retroCards.filter(card => !card.isEditable)
    }
  }
  ws.send(JSON.stringify(newState));
}