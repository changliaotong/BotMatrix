// 图表绘制工具类
class ChartUtil {
  // 绘制网格
  static drawGrid(ctx, width, height, gridColor = '#f0f0f0') {
    ctx.strokeStyle = gridColor;
    ctx.lineWidth = 1;
    
    // 水平网格线
    for (let i = 1; i < 5; i++) {
      const y = (height / 5) * i;
      ctx.beginPath();
      ctx.moveTo(0, y);
      ctx.lineTo(width, y);
      ctx.stroke();
    }
    
    // 垂直网格线
    for (let i = 1; i < 5; i++) {
      const x = (width / 5) * i;
      ctx.beginPath();
      ctx.moveTo(x, 0);
      ctx.lineTo(x, height);
      ctx.stroke();
    }
  }

  // 绘制折线图
  static drawLineChart(canvasId, data, color, label, onComplete) {
    const query = wx.createSelectorQuery();
    query.select(`#${canvasId}`).fields({ node: true, size: true }).exec((res) => {
      if (!res[0]) return;
      
      const canvas = res[0].node;
      const ctx = canvas.getContext('2d');
      const width = res[0].width;
      const height = res[0].height;
      
      // 设置canvas尺寸
      canvas.width = width;
      canvas.height = height;
      
      // 清空画布
      ctx.clearRect(0, 0, width, height);
      
      if (data.length === 0) return;
      
      // 绘制网格
      this.drawGrid(ctx, width, height);
      
      // 绘制数据线
      ctx.strokeStyle = color;
      ctx.lineWidth = 3;
      ctx.lineJoin = 'round';
      ctx.lineCap = 'round';
      
      const maxValue = Math.max(...data, 100);
      const xStep = width / (data.length - 1);
      const yScale = (height - 40) / maxValue;
      
      ctx.beginPath();
      data.forEach((value, index) => {
        const x = index * xStep;
        const y = height - 20 - (value * yScale);
        
        if (index === 0) {
          ctx.moveTo(x, y);
        } else {
          ctx.lineTo(x, y);
        }
      });
      ctx.stroke();
      
      // 绘制数据点
      ctx.fillStyle = color;
      data.forEach((value, index) => {
        const x = index * xStep;
        const y = height - 20 - (value * yScale);
        
        ctx.beginPath();
        ctx.arc(x, y, 4, 0, 2 * Math.PI);
        ctx.fill();
      });
      
      // 绘制当前值
      if (data.length > 0) {
        const currentValue = data[data.length - 1];
        ctx.fillStyle = color;
        ctx.font = 'bold 24rpx sans-serif';
        ctx.textAlign = 'right';
        ctx.fillText(`${Math.round(currentValue)}%`, width - 10, 30);
      }

      if (onComplete) onComplete();
    });
  }

  // 绘制网络流量图表
  static drawNetworkChart(canvasId, data, onComplete) {
    const query = wx.createSelectorQuery();
    query.select(`#${canvasId}`).fields({ node: true, size: true }).exec((res) => {
      if (!res[0]) return;
      
      const canvas = res[0].node;
      const ctx = canvas.getContext('2d');
      const width = res[0].width;
      const height = res[0].height;
      
      canvas.width = width;
      canvas.height = height;
      ctx.clearRect(0, 0, width, height);
      
      if (data.length === 0) return;
      
      this.drawGrid(ctx, width, height);
      
      const maxValue = Math.max(...data.map(d => Math.max(d.in, d.out)), 1000);
      const xStep = width / (data.length - 1);
      const yScale = (height - 40) / maxValue;
      
      // 绘制入站流量
      ctx.strokeStyle = '#07c160';
      ctx.lineWidth = 3;
      ctx.beginPath();
      data.forEach((point, index) => {
        const x = index * xStep;
        const y = height - 20 - (point.in * yScale);
        if (index === 0) ctx.moveTo(x, y);
        else ctx.lineTo(x, y);
      });
      ctx.stroke();
      
      // 绘制出站流量
      ctx.strokeStyle = '#1989fa';
      ctx.beginPath();
      data.forEach((point, index) => {
        const x = index * xStep;
        const y = height - 20 - (point.out * yScale);
        if (index === 0) ctx.moveTo(x, y);
        else ctx.lineTo(x, y);
      });
      ctx.stroke();
      
      // 绘制图例
      ctx.fillStyle = '#07c160';
      ctx.fillRect(width - 150, 10, 20, 3);
      ctx.fillStyle = '#333';
      ctx.font = '22rpx sans-serif';
      ctx.textAlign = 'left';
      ctx.fillText('入站', width - 120, 20);
      
      ctx.fillStyle = '#1989fa';
      ctx.fillRect(width - 80, 10, 20, 3);
      ctx.fillStyle = '#333';
      ctx.fillText('出站', width - 50, 20);

      if (onComplete) onComplete();
    });
  }

  // 绘制环形图
  static drawRingChart(canvasId, data, colors, onComplete) {
    const query = wx.createSelectorQuery();
    query.select(`#${canvasId}`).fields({ node: true, size: true }).exec((res) => {
      if (!res[0]) return;
      
      const canvas = res[0].node;
      const ctx = canvas.getContext('2d');
      const width = res[0].width;
      const height = res[0].height;
      const centerX = width / 2;
      const centerY = height / 2;
      const radius = Math.min(width, height) / 3;
      const innerRadius = radius * 0.6;
      
      canvas.width = width;
      canvas.height = height;
      ctx.clearRect(0, 0, width, height);
      
      if (!data || data.length === 0) return;
      
      const total = data.reduce((sum, item) => sum + item.value, 0);
      let currentAngle = -Math.PI / 2;
      
      // 绘制环形
      data.forEach((item, index) => {
        const angle = (item.value / total) * Math.PI * 2;
        ctx.fillStyle = colors[index % colors.length];
        ctx.beginPath();
        ctx.arc(centerX, centerY, radius, currentAngle, currentAngle + angle);
        ctx.arc(centerX, centerY, innerRadius, currentAngle + angle, currentAngle, true);
        ctx.closePath();
        ctx.fill();
        
        currentAngle += angle;
      });
      
      // 绘制中心文字
      ctx.fillStyle = '#333';
      ctx.font = 'bold 36rpx sans-serif';
      ctx.textAlign = 'center';
      ctx.textBaseline = 'middle';
      ctx.fillText(`${Math.round(total)}%`, centerX, centerY);

      if (onComplete) onComplete();
    });
  }
}

module.exports = ChartUtil;