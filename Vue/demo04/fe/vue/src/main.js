import Go from './wasm_exec'

import Vue from 'vue'
import App from './App.vue'

/* eslint no-undef: "off"*/
if (!WebAssembly.instantiateStreaming) { // polyfill
  WebAssembly.instantiateStreaming = async (resp, importObject) => {
      const source = await (await resp).arrayBuffer();
      return await WebAssembly.instantiate(source, importObject);
  };
}
const go = new Go()
WebAssembly.instantiateStreaming(
  fetch("calc.wasm"), go.importObject)
  // .then(async (result) => {
  // await go.run(result.instance)
  .then((result) => {
    go.run(result.instance)
    // console.log(waAdd(...Array("2", "1")))

    // 全局注册
    Vue.prototype.$go = {
      add: waAdd,
      sub: waSub,
      multi: waMulti,
      divi: waDivi,
    }
  })

new Vue({
  render: h => h(App)
}).$mount('#app')
