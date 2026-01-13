import { createApp } from 'vue'
import { createPinia } from 'pinia'
import router from './router/index'
import i18n from './utils/i18n'
import './style.css'
import App from './App.vue'

const app = createApp(App)
app.use(createPinia())
app.use(i18n)
app.use(router)
app.mount('#app')
