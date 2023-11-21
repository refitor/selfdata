// main.js
import Vue from 'vue'
import ViewUI from 'view-design';
import 'view-design/dist/styles/iview.css';
import App from './App.vue';
import router from './router';
import VueI18n from 'vue-i18n';
import cookie from 'vue-cookie';
import store from './store';
import axios from 'axios';
import formCreate from '@form-create/iview'

Vue.use(VueI18n);
Vue.use(ViewUI);
Vue.use(formCreate)
// Vue.locale = () => {};
Vue.config.productionTip = false
Vue.prototype.$cookie = cookie;

// axios with cookie
axios.defaults.withCredentials=true;
Vue.prototype.$axios = axios;

// ignore console.log
// console.log = ()=>{}

const i18n = new VueI18n({
    locale: localStorage.getItem('lang') !== null ? localStorage.getItem('lang'):'zh-CN',    // 语言标识, 通过切换locale的值来实现语言切换,this.$i18n.locale 
    messages: {
      'zh-CN': require('./lang/zh'),   // 中文语言包
      'en-US': require('./lang/en')    // 英文语言包
    }
})

new Vue({
  i18n,
  store,
  router,
  render: h => h(App)
}).$mount('#app')

// // init public key
// if (store.state.PublicKey === '') {
//   axios.get('/api/cert/public')
//   .then(function (response) {
//     if (store.state.AESKey === '') store.commit('Set_AESKey', {AESKey: Crypt.generatekey(16)});
//     store.commit('Set_PublicKey', {PublicKey: response.data});
  
//     new Vue({
//       i18n,
//       store,
//       router,
//       render: h => h(App)
//     }).$mount('#app')
//   })
//   .catch(function (response) {
//       console.log(response);
//   });
// }

// // init ajax-hook
// proxy({
//   //请求发起前进入
//   onRequest: (config, handler) => {
//       if (config.body !== undefined && config.body !== null) {
//         if (typeof(config.body) === 'string' && config.body.substr(0, 7) == 'SFv3rg:') {
//           config.body = 'SFv3rg:' + Crypt.aesEncrypt(config.body.split('SFv3rg:')[1], store.state.AESKey);
//         }
//       }                 
//       config.headers.SFv3rg = Crypt.rsaEncrypt(store.state.AESKey, store.state.PublicKey);
//       console.log('===request===>', config)
//       handler.next(config);
//   },
//   //请求发生错误时进入，比如超时；注意，不包括http状态码错误，如404仍然会认为请求成功
//   onError: (err, handler) => {
//       console.log(err.type)
//       handler.next(err)
//   },
//   //请求成功后进入
//   onResponse: (response, handler) => {
//       console.log('===response before===>', response)
//       if (typeof(response.response) === 'string') {
//         if (response.response.substr(0, 7) == 'SFv3rg:') {
//           let datas = response.response.split('SFv3rg:');
//           if (response.config.url.indexOf('/api/cert/public') > -1)
//               response.response = Crypt.DecodeByBase64(datas[1]);
//           else
//           {
//               let strDecode = Crypt.aesDecrypt(datas[1], store.state.AESKey);
//               if (typeof(strDecode) === 'string') {
//                 response.response = strDecode;
//               } else {
//                 response.response = JSON.parse(strDecode);
//               }
//           }
//         }
//       }
//       console.log('===response after===>', response)
//       handler.next(response)
//   }
// })
