import { createI18n } from 'vue-i18n'
const messages = {
  zh: {
    nav: { home: '首页', products: '产品', pricing: '价格', about: '关于', contact: '联系' },
    home: { title: '机器人早喵', subtitle: '智能协作助手，提升工作效率', cta: '了解产品' },
    products: { title: '产品列表', view: '查看详情' },
    pricing: { title: '价格计划', contact: '联系咨询' },
    about: { title: '关于我们', story: '品牌故事与愿景' },
    contact: { title: '联系我们', name: '姓名', email: '邮箱', message: '信息', agree: '同意隐私政策', submit: '提交' },
    product: { contact: '联系经销商' },
    blog: { title: '博客', note: '后续内容' }
  },
  en: {
    nav: { home: 'Home', products: 'Products', pricing: 'Pricing', about: 'About', contact: 'Contact' },
    home: { title: 'Robot Meow', subtitle: 'Intelligent collaboration assistant to boost productivity', cta: 'Learn More' },
    products: { title: 'Product List', view: 'View Details' },
    pricing: { title: 'Pricing Plans', contact: 'Contact Us' },
    about: { title: 'About Us', story: 'Brand story & vision' },
    contact: { title: 'Contact Us', name: 'Name', email: 'Email', message: 'Message', agree: 'Agree to privacy policy', submit: 'Submit' },
    product: { contact: 'Contact Distributor' },
    blog: { title: 'Blog', note: 'More posts coming soon' }
  }
}
export default createI18n({ locale: 'zh', fallbackLocale: 'zh', messages })
