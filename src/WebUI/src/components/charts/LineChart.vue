<script setup lang="ts">
import { computed } from 'vue';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler,
  type ChartOptions,
  type ChartData
} from 'chart.js';
import { Line } from 'vue-chartjs';

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
);

const props = withDefaults(defineProps<{
  data?: number[] | any[];
  labels?: string[];
  color?: string;
  fill?: boolean;
}>(), {
  data: () => [],
  labels: undefined,
  color: '#10b981', // Changed from #00ff41 to match emerald matrix color
  fill: true
});

// Helper to convert hex to rgba
const getRgba = (hex: string, alpha: number) => {
  if (!hex || !hex.startsWith('#')) return `rgba(16, 185, 129, ${alpha})`;
  try {
    const r = parseInt(hex.slice(1, 3), 16);
    const g = parseInt(hex.slice(3, 5), 16);
    const b = parseInt(hex.slice(5, 7), 16);
    return `rgba(${r}, ${g}, ${b}, ${alpha})`;
  } catch (e) {
    return `rgba(16, 185, 129, ${alpha})`;
  }
};

const chartData = computed<ChartData<'line'>>(() => {
  const baseColor = props.color || '#10b981';
  return {
    labels: props.labels || (props.data || []).map((_, i) => i.toString()),
    datasets: [
      {
        label: '',
        data: props.data || [],
        borderColor: baseColor,
        backgroundColor: props.fill ? getRgba(baseColor, 0.1) : 'transparent',
        fill: props.fill,
        tension: 0.4,
        pointRadius: 0,
        borderWidth: 2,
      },
    ],
  };
});

const chartOptions = computed<ChartOptions<'line'>>(() => ({
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      display: false,
    },
    tooltip: {
      enabled: true,
      mode: 'index',
      intersect: false,
    },
  },
  scales: {
    x: {
      display: false,
    },
    y: {
      display: false,
      beginAtZero: false,
      grace: '10%'
    },
  },
}));
</script>

<template>
  <div class="w-full h-full">
    <Line :data="chartData" :options="chartOptions" />
  </div>
</template>
