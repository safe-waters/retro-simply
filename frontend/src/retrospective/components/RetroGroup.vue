<template>
  <div>
    <div class="d-flex">
      <i
        class="fa fa-align-justify handle text-dark ps-2 pt-1"
        :style="{ cursor: isDraggable ? 'pointer' : '' }"
        style="height: 0px"
      ></i>
      <i
        v-if="group.id !== 'default'"
        class="fas mt-1 ps-1 text-dark"
        :class="{
          'fa-compress-alt': show,
          'fa-expand-alt': !show,
          grow: canToggleShow,
        }"
        :style="{
          cursor: canToggleShow ? 'pointer' : '',
        }"
        style="height: 0px"
        @click="toggleShow"
      ></i>
      <div
        class="text-dark text-break ms-2 me-1 mb-1 mt-0"
        :contenteditable="group.isEditable"
        :ref="group.id"
        placeholder="Type to add group title..."
        style="white-space: pre-wrap; outline: none; flex-grow: 1"
        v-text="group.title"
        @paste.prevent
        @paste="paste"
        @keydown.enter.prevent
        @keyup.enter="updateGroupTitle($event.target.innerText)"
      ></div>
      <span
        v-if="group.id !== 'default'"
        class="me-2 text-dark"
        style="white-space: nowrap"
        >votes:&nbsp;<strong>{{ groupVotes }}</strong></span
      >
    </div>
    <div v-show="show">
      <draggable
        :disabled="!isDraggable"
        handle=".handle"
        v-model="draggableRetroCards"
        :group="group.columnId"
      >
        <retro-card
          v-for="retroCard in draggableRetroCards"
          v-show="!retroCard.isDeleted"
          class="mb-1"
          :key="retroCard.id"
          :retroCard="retroCard"
          :cardStyle="cardStyle"
          :isDraggable="isDraggable"
        />
      </draggable>
    </div>
  </div>
</template>

<script>
import draggable from "vuedraggable";
import RetroCard from "./RetroCard.vue";

export default {
  name: "RetroGroup",
  props: ["group", "cardStyle", "isDraggable"],
  mounted: function () {
    if (this.group.isEditable) {
      this.$refs[this.group.id].focus();
    }
  },
  data: function () {
    return {
      show: true,
    };
  },
  computed: {
    draggableRetroCards: {
      get() {
        return JSON.parse(JSON.stringify(this.group.retroCards));
      },
      set(newCards) {
        // move card
        if (newCards.length === this.group.retroCards.length) {
          let payload = {
            group: this.group,
            newRetroCards: newCards,
          };

          this.$store.commit("switchCardInSameGroup", payload);
          return;
        }

        // a new card has been added
        if (
          newCards.filter((card) => !card.isDeleted).length >
          this.group.retroCards.filter((card) => !card.isDeleted).length
        ) {
          let payload = {
            group: this.group,
            newRetroCards: newCards,
            send: true,
          };
          this.$store.commit("switchCardGroup", payload);

          return;
        }
      },
    },
    groupVotes: function () {
      let votes = 0;
      for (let i = 0; i < this.group.retroCards.length; i++) {
        if (!this.group.retroCards[i].isDeleted) {
          votes += this.group.retroCards[i].numVotes;
        }
      }
      return votes;
    },
    canToggleShow: function () {
      return this.group.retroCards.length > 0 || !this.show;
    },
  },
  methods: {
    updateGroupTitle: function (title) {
      if (title.trim() !== "") {
        let newGroup = {
          id: this.group.id,
          columnId: this.group.columnId,
          isEditable: false,
          title: title,
          retroCards: this.group.retroCards,
        };

        let payload = {
          send: true,
          group: newGroup,
        };
        this.$store.commit("updateGroupTitle", payload);
      }
    },
    toggleShow: function () {
      if (this.canToggleShow) {
        this.show = !this.show;
      }
    },
    paste: function (e) {
      let pastedText = e.clipboardData.getData("Text");
      if (pastedText) {
        document.execCommand("insertText", false, pastedText);
      }
    },
  },
  components: {
    "retro-card": RetroCard,
    draggable: draggable,
  },
};
</script>

<style scoped>
.grow {
  transition: all 0.1s;
}

.grow:hover {
  transform: scale(1.02);
}

.grow:active {
  transform: scale(1.02);
  opacity: 0.7;
}

[contenteditable][placeholder]:empty:before {
  content: attr(placeholder);
}
</style>