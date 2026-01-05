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
  color: '#00ff41',
  fill: true
});

const chartData = computed<ChartData<'line'>>(() => ({
  labels: props.labels || (props.data || []).map((_, i) => i.toString()),
  datasets: [
    {
      label: '',
      data: props.data || [],
      borderColor: props.color || '#00ff41',
      backgroundColor: props.fill ? (props.color ? `${props.color}20` : 'rgba(0, 255, 65, 0.1)') : 'transparent',
      fill: props.fill,
      tension: 0.4,
      pointRadius: 0,
      borderWidth: 2,
    },
  ],
}));

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
