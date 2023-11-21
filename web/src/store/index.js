import Vue from 'vue'
import Vuex from 'vuex'
Vue.use(Vuex)

const store = new Vuex.Store({
　　 state:{
        AESKey:'',
        PublicKey: ''
　　 },
    getters: {
        AESKey: state => state.AESKey,
        PublicKey: state => state.PublicKey,
    },
    mutations: {
        Set_AESKey(state, data) {
            state.AESKey = data.AESKey;
            
        },
        Set_PublicKey(state, data) {
            state.PublicKey = data.PublicKey;
        },
    },
})
export default store