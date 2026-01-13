/** @type {import('tailwindcss').Config} */
export default {
  darkMode: 'class',
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        matrix: '#00ff41',
        meow: {
          honey: '#FFB347',  // 温暖金蜜
          peach: '#FF9A8B',  // 柔和蜜桃
          rose: '#E0607E',   // 优雅玫瑰
          cream: '#FFFBF5',  // 纯净奶油
          dark: '#4A4A4A',   // 暖调深灰
          gold: '#D4AF37',   // 尊贵金
          lavender: '#E6E6FA', // 熏衣草
        },
        cyber: {
          black: '#050508',
          dark: '#0A0A0F',
          surface: '#12121A',
          neon: '#00F2FF',
          pink: '#FF00E5',
          green: '#39FF14',
          yellow: '#FFF01F',
          gray: '#2D2D3F',
          border: 'rgba(255, 255, 255, 0.1)',
        }
      },
      backgroundImage: {
        'cyber-gradient': 'linear-gradient(135deg, #0A0A0F 0%, #12121A 100%)',
        'meow-gradient': 'linear-gradient(135deg, #FFB347 0%, #FF9A8B 100%)',
        'neon-glow': 'radial-gradient(circle, rgba(0, 242, 255, 0.15) 0%, transparent 70%)',
      },
      animation: {
        'float': 'float 6s ease-in-out infinite',
        'pulse-slow': 'pulse 4s cubic-bezier(0.4, 0, 0.6, 1) infinite',
        'glow': 'glow 2s ease-in-out infinite alternate',
        'scroll': 'scroll 40s linear infinite',
        'spin-slow': 'spin 12s linear infinite',
        'reverse-spin': 'reverse-spin 15s linear infinite',
        'gradient': 'gradient 8s linear infinite',
      },
      keyframes: {
        gradient: {
          '0%': { backgroundPosition: '0% 50%' },
          '50%': { backgroundPosition: '100% 50%' },
          '100%': { backgroundPosition: '0% 50%' },
        },
        float: {
          '0%, 100%': { transform: 'translateY(0)' },
          '50%': { transform: 'translateY(-20px)' },
        },
        glow: {
          '0%': { boxShadow: '0 0 5px rgba(0, 242, 255, 0.2), 0 0 10px rgba(0, 242, 255, 0.2)' },
          '100%': { boxShadow: '0 0 20px rgba(0, 242, 255, 0.6), 0 0 30px rgba(0, 242, 255, 0.4)' },
        },
        scroll: {
          '0%': { transform: 'translateX(0)' },
          '100%': { transform: 'translateX(-50%)' },
        },
        'reverse-spin': {
          from: { transform: 'rotate(360deg)' },
          to: { transform: 'rotate(0deg)' },
        }
      }
    },
  },
  plugins: [],
}
