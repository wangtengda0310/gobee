import { createApp } from 'vue'
import App, {router} from './App.vue'
import naive from "naive-ui";

createApp(App).use(router).use(naive).mount('#app')