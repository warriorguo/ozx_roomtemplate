import type { Template } from '../types/newTemplate';

/**
 * 基于Ground图层生成PNG缩略图
 * @param template 模板数据
 * @param size 缩略图尺寸（正方形）
 * @returns Promise<string> Base64编码的PNG数据URL
 */
export function generateThumbnail(template: Template, size: number = 120): Promise<string> {
  return new Promise((resolve, reject) => {
    try {
      // 创建canvas
      const canvas = document.createElement('canvas');
      const ctx = canvas.getContext('2d');
      
      if (!ctx) {
        reject(new Error('Unable to get canvas context'));
        return;
      }
      
      canvas.width = size;
      canvas.height = size;
      
      // 计算缩放比例
      const scaleX = size / template.width;
      const scaleY = size / template.height;
      const scale = Math.min(scaleX, scaleY);
      
      // 计算居中位置
      const scaledWidth = template.width * scale;
      const scaledHeight = template.height * scale;
      const offsetX = (size - scaledWidth) / 2;
      const offsetY = (size - scaledHeight) / 2;
      
      // 设置背景色（浅灰色）
      ctx.fillStyle = '#f0f0f0';
      ctx.fillRect(0, 0, size, size);
      
      // 绘制Ground图层
      ctx.fillStyle = '#90EE90'; // Ground的绿色
      
      for (let y = 0; y < template.height; y++) {
        for (let x = 0; x < template.width; x++) {
          if (template.ground[y][x] === 1) {
            const pixelX = offsetX + x * scale;
            const pixelY = offsetY + y * scale;
            ctx.fillRect(pixelX, pixelY, scale, scale);
          }
        }
      }
      
      // 添加边框
      ctx.strokeStyle = '#cccccc';
      ctx.lineWidth = 1;
      ctx.strokeRect(offsetX - 0.5, offsetY - 0.5, scaledWidth + 1, scaledHeight + 1);
      
      // 转换为PNG DataURL
      const dataURL = canvas.toDataURL('image/png');
      resolve(dataURL);
      
    } catch (error) {
      reject(error);
    }
  });
}

/**
 * 生成更详细的缩略图，包含多个图层
 * @param template 模板数据
 * @param size 缩略图尺寸
 * @returns Promise<string> Base64编码的PNG数据URL
 */
export function generateDetailedThumbnail(template: Template, size: number = 120): Promise<string> {
  return new Promise((resolve, reject) => {
    try {
      const canvas = document.createElement('canvas');
      const ctx = canvas.getContext('2d');
      
      if (!ctx) {
        reject(new Error('Unable to get canvas context'));
        return;
      }
      
      canvas.width = size;
      canvas.height = size;
      
      // 计算缩放比例和位置
      const scaleX = size / template.width;
      const scaleY = size / template.height;
      const scale = Math.min(scaleX, scaleY);
      
      const scaledWidth = template.width * scale;
      const scaledHeight = template.height * scale;
      const offsetX = (size - scaledWidth) / 2;
      const offsetY = (size - scaledHeight) / 2;
      
      // 设置背景色
      ctx.fillStyle = '#f8f9fa';
      ctx.fillRect(0, 0, size, size);
      
      // 图层颜色定义
      const layerColors = {
        ground: '#90EE90',    // 浅绿色
        static: '#FFA500',    // 橙色  
        turret: '#4169E1',    // 蓝色
        mobGround: '#FFD700', // 黄色
        mobAir: '#87CEEB',    // 天蓝色
      };
      
      // 按优先级绘制图层（后绘制的在上层）
      const layers: Array<keyof typeof layerColors> = ['ground', 'static', 'turret', 'mobGround', 'mobAir'];
      
      for (const layerName of layers) {
        ctx.fillStyle = layerColors[layerName];
        
        for (let y = 0; y < template.height; y++) {
          for (let x = 0; x < template.width; x++) {
            if (template[layerName][y][x] === 1) {
              const pixelX = offsetX + x * scale;
              const pixelY = offsetY + y * scale;
              ctx.fillRect(pixelX, pixelY, scale, scale);
            }
          }
        }
      }
      
      // 添加网格线（如果格子足够大）
      if (scale >= 4) {
        ctx.strokeStyle = '#ffffff40';
        ctx.lineWidth = 0.5;
        
        // 垂直线
        for (let x = 0; x <= template.width; x++) {
          const lineX = offsetX + x * scale;
          ctx.beginPath();
          ctx.moveTo(lineX, offsetY);
          ctx.lineTo(lineX, offsetY + scaledHeight);
          ctx.stroke();
        }
        
        // 水平线
        for (let y = 0; y <= template.height; y++) {
          const lineY = offsetY + y * scale;
          ctx.beginPath();
          ctx.moveTo(offsetX, lineY);
          ctx.lineTo(offsetX + scaledWidth, lineY);
          ctx.stroke();
        }
      }
      
      // 添加边框
      ctx.strokeStyle = '#dee2e6';
      ctx.lineWidth = 2;
      ctx.strokeRect(offsetX - 1, offsetY - 1, scaledWidth + 2, scaledHeight + 2);
      
      const dataURL = canvas.toDataURL('image/png');
      resolve(dataURL);
      
    } catch (error) {
      reject(error);
    }
  });
}