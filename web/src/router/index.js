import Vue from 'vue'
import Auth from '../pages/auth.vue'
import Conf from '../pages/conf.vue'
import Home from '../pages/home.vue'
import VueRouter from 'vue-router'
import {webResponse} from '../toolchain/help.js';
Vue.use(VueRouter)

const routes = [
  {
    path: '/',
    name: 'home',
    component: Home,
    meta: { 
      requireAuth: true
    },
  },
  {
    path: '/conf',
    name: 'conf',
    component: Conf
  },
  {
    path: '/auth',
    name: 'auth',
    component: Auth
  },
]

const router = new VueRouter({
  routes
})

// // 判断是否需要登录权限 以及是否登录
// router.beforeEach((to, from, next) => {
//   if (to.path !== '/auth' && to.matched.some(res => res.meta.requireAuth)) {// 判断是否需要登录权限
//     let redirect = to.fullPath.indexOf('#') === -1 ? '/#' + to.fullPath : to.fullPath;
//     if (localStorage.getItem('isLogin') === 'true') {
//       next()
//     } else {
//       next({
//         path: '/auth',
//         query: {redirectUrl: redirect}
//       })
//     }
//   } else {
//     next()
//   }
// })
export default router
