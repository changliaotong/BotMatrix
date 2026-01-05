<script setup lang="ts">
import { onMounted, ref, onUnmounted } from 'vue';
import { useSystemStore } from '@/stores/system';

const canvasRef = ref<HTMLCanvasElement | null>(null);
const systemStore = useSystemStore();

onMounted(() => {
  const canvas = canvasRef.value;
  if (!canvas) return;

  const ctx = canvas.getContext('2d', { alpha: false }); // Disable alpha for performance
  if (!ctx) return;

  let animationFrameId: number;

  const updateSize = () => {
    canvas.width = window.innerWidth;
    canvas.height = window.innerHeight;
  };
  updateSize();

  // Optimized character set: Focus on high-impact symbols
  const chars = '0123456789ABCDEFHIJKLMNOPQRSTUVWXYZｦｧｨｩｪｫｬｭｮｯｰｱｲｳｴｵｶｷｸｹｺｻｼｽｾｿﾀﾁﾂﾃﾄﾅﾆﾇﾈﾉﾊﾋﾌﾍﾎﾏﾐﾑﾒﾓﾔﾕﾖﾗﾘﾙﾚﾛﾜﾝ';
  const fontSize = 14;
  const columns = Math.ceil(canvas.width / fontSize);

  // Advanced Drop Object
  class Drop {
    x: number;
    y: number;
    speed: number = 0;
    length: number = 0;
    lastChar: string = '';
    chars: string[] = [];
    updateCount: number = 0;

    constructor(x: number) {
      this.x = x;
      this.reset();
      this.y = Math.random() * -100; // Random start offset
    }

    reset() {
      this.y = 0;
      this.speed = 1.5 + Math.random() * 4;
      this.length = 15 + Math.floor(Math.random() * 25);
      this.lastChar = '';
      this.updateCount = 0;
    }

    draw() {
      this.updateCount++;
      
      // Draw head (white/glowing)
      const headChar = chars[Math.floor(Math.random() * chars.length)];
      ctx!.font = `bold ${fontSize}px monospace`;
      ctx!.fillStyle = '#ffffff';
      ctx!.fillText(headChar, this.x, this.y * fontSize);

      // Draw tail
      for (let i = 1; i < this.length; i++) {
        const charY = (this.y - i) * fontSize;
        if (charY < 0 || charY > canvas!.height) continue;

        const opacity = 1 - (i / this.length);
        const char = chars[Math.floor(Math.random() * chars.length)];
        
        // Varying green intensity
        const green = Math.floor(255 * opacity);
        ctx!.fillStyle = `rgb(0, ${green}, 0)`;
        ctx!.fillText(char, this.x, charY);
      }

      this.y += this.speed * 0.15;

      if (this.y * fontSize > canvas!.height + (this.length * fontSize)) {
        this.reset();
      }
    }
  }

  const drops: Drop[] = [];
  for (let i = 0; i < columns; i++) {
    drops.push(new Drop(i * fontSize));
  }

  const render = () => {
    // Clear with extreme persistence (low alpha fill)
    ctx!.fillStyle = 'rgba(0, 0, 0, 0.12)';
    ctx!.fillRect(0, 0, canvas.width, canvas.height);

    drops.forEach(drop => drop.draw());
    animationFrameId = requestAnimationFrame(render);
  };

  render();

  window.addEventListener('resize', updateSize);

  onUnmounted(() => {
    cancelAnimationFrame(animationFrameId);
    window.removeEventListener('resize', updateSize);
  });
});
</script>

<template>
  <canvas ref="canvasRef" class="matrix-rain-extreme"></canvas>
</template>

<style scoped>
.matrix-rain-extreme {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
  z-index: -1;
  background: #000;
  opacity: 0.8; /* Higher density visual */
}
</style>
