<template>
  <div class="rounded pb-3" :class="cardStyle.backgroundColor">
    <div class="d-flex align-items-end flex-column me-2">
      <div>
        <span class="text-white">
          <i
            @click="upVote"
            class="fas fa-thumbs-up p-2"
            :class="{ 'grow-upvote': isUpVotable }"
            :style="{ cursor: isUpVotable ? 'pointer' : '' }"
          ></i
          ><span>{{ retroCard.numVotes }}</span>
        </span>
      </div>
    </div>

    <div class="d-flex">
      <i
        class="fa fa-align-justify handle text-white ps-2 pt-1"
        :style="{ cursor: isDraggable ? 'pointer' : '' }"
        style="height: 0px"
      ></i>
      <div
        class="text-break ps-3 pe-3 text-white"
        :class="cardStyle.backgroundColor"
        :ref="retroCard.id"
        style="white-space: pre-wrap; outline: none"
        :contenteditable="retroCard.isEditable"
        v-text="retroCard.message"
        @paste.prevent
        @paste="paste"
        @keydown.enter="updateMessage"
        @keyup.enter="updateMessage"
        placeholder="Type to add new item..."
      ></div>
    </div>
  </div>
</template>

<script>
import { mapGetters } from "vuex";

export default {
  name: "RetroCard",
  props: ["retroCard", "cardStyle", "isDraggable"],
  mounted: function () {
    if (this.retroCard.isEditable) {
      this.$refs[this.retroCard.id].focus();
    }
  },
  methods: {
    upVote: function () {
      if (this.isUpVotable) {
        let card = {
          id: this.retroCard.id,
          columnId: this.retroCard.columnId,
          message: this.retroCard.message,
          numVotes: this.retroCard.numVotes + 1,
          isEditable: this.retroCard.isEditable,
          groupId: this.retroCard.groupId,
          isDeleted: this.retroCard.isDeleted,
          lastModified: Date.now(),
        };

        let payload = {
          card: card,
          action: {
            title: "upVote",
            oldCard: JSON.parse(JSON.stringify(this.retroCard)),
            newCard: JSON.parse(JSON.stringify(card)),
          },
          send: true,
        };

        this.$store.commit("updateRetroCard", payload);
      }
    },
    updateMessage: function (e) {
      if (e.shiftKey && e.key === "Enter") {
        return;
      }

      e.preventDefault();

      let message = e.target.innerText.trim();
      if (message !== "" && this.connected) {
        let card = {
          id: this.retroCard.id,
          columnId: this.retroCard.columnId,
          message: message,
          numVotes: this.retroCard.numVotes,
          isEditable: false,
          groupId: this.retroCard.groupId,
          isDeleted: this.retroCard.isDeleted,
          lastModified: Date.now(),
        };

        let payload = {
          card: card,
          action: null,
          send: true,
        };

        this.$store.commit("updateRetroCard", payload);
      }
    },
    paste: function (e) {
      let pastedText = e.clipboardData.getData("Text");
      if (pastedText) {
        document.execCommand("insertText", false, pastedText);
      }
    },
  },
  computed: {
    isUpVotable: function () {
      return !this.retroCard.isEditable && this.connected;
    },
    ...mapGetters(["connected"]),
  },
};
</script>

<style scoped>
.grow-upvote {
  transition: all 0.1s;
}

.grow-upvote:hover {
  transform: scale(1.08);
}

.grow-upvote:active {
  transform: scale(1.08);
  opacity: 0.7;
}

[contenteditable][placeholder]:empty:before {
  content: attr(placeholder);
}
</style>