<script setup lang="ts">
import { onMounted, ref } from 'vue';

const particles = ref<any[]>([]);

onMounted(() => {
  const count = 20;
  for (let i = 0; i < count; i++) {
    particles.value.push({
      id: i,
      x: Math.random() * 100,
      y: Math.random() * 100,
      size: Math.random() * 20 + 10,
      duration: Math.random() * 5 + 3,
      delay: Math.random() * 5,
      type: Math.random() > 0.5 ? '♥' : '✨'
    });
  }
});
</script>

<template>
  <div class="kawaii-sparkles">
    <div 
      v-for="p in particles" 
      :key="p.id" 
      class="particle"
      :style="{
        left: p.x + '%',
        top: p.y + '%',
        fontSize: p.size + 'px',
        animationDuration: p.duration + 's',
        animationDelay: p.delay + 's'
      }"
    >
      {{ p.type }}
    </div>
  </div>
</template>

<style scoped>
.kawaii-sparkles {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
  z-index: 0;
  overflow: hidden;
}

.particle {
  position: absolute;
  opacity: 0;
  color: #ffb6c1;
  animation: float-up linear infinite;
  text-shadow: 0 0 10px rgba(255, 182, 193, 0.5);
}

@keyframes float-up {
  0% {
    transform: translateY(100vh) rotate(0deg) scale(0.5);
    opacity: 0;
  }
  10% {
    opacity: 0.6;
  }
  90% {
    opacity: 0.6;
  }
  100% {
    transform: translateY(-100px) rotate(360deg) scale(1.2);
    opacity: 0;
  }
}
</style>
