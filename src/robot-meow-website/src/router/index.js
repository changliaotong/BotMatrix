import { createRouter, createWebHistory } from 'vue-router'
import Home from '../pages/Home.vue'
import Products from '../pages/Products.vue'
import ProductDetail from '../pages/ProductDetail.vue'
import Pricing from '../pages/Pricing.vue'
import About from '../pages/About.vue'
import Contact from '../pages/Contact.vue'
import Blog from '../pages/Blog.vue'

const routes = [
  { path: '/', component: Home },
  { path: '/products', component: Products },
  { path: '/products/:slug', component: ProductDetail },
  { path: '/pricing', component: Pricing },
  { path: '/about', component: About },
  { path: '/contact', component: Contact },
  { path: '/blog', component: Blog },
  { path: '/en', redirect: '/' },
  { path: '/:catchAll(.*)', redirect: '/' }
]

const router = createRouter({ history: createWebHistory(), routes })

export default router
