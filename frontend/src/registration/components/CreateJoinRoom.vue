<template>
  <div class="row">
    <div class="col-8 offset-2">
      <h1 class="display-3 text-center">{{ title }}</h1>
      <form @submit.prevent="submit">
        <div class="mb-3">
          <label class="form-label">Room&nbsp;Id</label>
          <input :value="room.id" @input="setId" class="form-control" />
        </div>
        <div class="mb-3">
          <label class="form-label">Password</label>
          <input
            :value="room.password"
            @input="setPassword"
            type="password"
            class="form-control"
          />
        </div>
        <div class="text-center">
          <button
            type="submit"
            class="btn btn-primary text-center"
            :disabled="disabled"
          >
            Submit
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script>
export default {
  name: "CreateJoinRoom",
  props: ["type", "title"],
  methods: {
    submit: function () {
      this.$store.dispatch("submit", this.type);
    },
    setId: function (e) {
      let payload = {
        id: e.target.value,
        type: this.type,
      };

      this.$store.commit("setId", payload);
    },
    setPassword: function (e) {
      let payload = {
        password: e.target.value,
        type: this.type,
      };

      this.$store.commit("setPassword", payload);
    },
  },
  computed: {
    disabled: function () {
      return this.room.id.trim() === "" || this.room.password.length === 0;
    },
    room: function () {
      return this.$store.state.data[this.type];
    },
  },
};
</script>