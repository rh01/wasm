<template>
  <div class="hello">
    <h1>WASM Demo 😂</h1>

    <input type="text" v-model="value1" />
    <span>{{ operator }}</span>
    <input type="text" v-model="value2" />
    =
    <input type="text" v-model="calcResult" disabled />
    <hr />
    <button @click="add()" id="addButton">Add</button>
    <button @click="sub()" id="subButton">Sub</button>
    <button @click="mul()" id="subButton">Mult</button>
    <button @click="div()" id="subButton">Div</button>
    <button @click="log()" id="subButton">Log</button>
    <button @click="asyncOne()" id="subButton">asyncOne</button>
    <button @click="fetchHttp()" id="subButton">fetchHttp</button>
    <button @click="fetchMongoDocument()" id="subButton">fetchMongoDocument</button>
  </div>
</template>

<script>
// /* eslint no-undef: "off"*/
// /* global waAdd, waSub, waMulti, waDivi */
export default {
  name: "HelloWorld",
  props: {
    msg: String,
  },
  data() {
    return {
      operator: "?",
      value1: "",
      value2: "",
      calcResult: "",
    };
  },
  created() {},
  mounted() {},
  methods: {
    async fetchMongoDocument() {
      try {
        const response = await this.$go.fetchMongoDocument("127.0.0.1");
        const message = await response.json();
        console.log(message);
      } catch (err) {
        console.error("Caught exception", err);
      }
    },
    async fetchHttp() {
      try {
        const response = await this.$go.fetchHttp("https://api.taylor.rest/");
        const message = await response.json();
        console.log(message);
      } catch (err) {
        console.error("Caught exception", err);
      }
    },
    async asyncOne() {
      let values = [this.value1, this.value2];

      try {
        console.log(await this.$go.asyncOne(...values));
      } catch (err) {
        console.error("Caught exception", err);
      }
    },
    async log() {
      let values = [this.value1, this.value2];
      try {
        console.log(await this.$go.log(...values));
      } catch (err) {
        console.error("Caught exception", err);
      }
    },
    add() {
      let values = [this.value1, this.value2];

      // // eslint-disable-next-line
      // this.calcResult = waAdd(...values)
      this.calcResult = this.$go.add(...values);
      this.operator = "+";
    },
    sub() {
      let values = [this.value1, this.value2];

      // this.calcResult = waSub(...values) // // eslint-disable-line
      this.calcResult = this.$go.sub(...values);
      this.operator = "-";
    },
    mul() {
      let values = [this.value1, this.value2];

      // this.calcResult = waMulti(...values)
      this.calcResult = this.$go.multi(...values);
      this.operator = "*";
    },
    div() {
      let values = [this.value1, this.value2];

      // this.calcResult = waDivi(...values)
      this.calcResult = this.$go.divi(...values);
      this.operator = "/";
    },
  },
};
</script>

<style scoped>
</style>
