import type { Template } from '../types/template';
import { validateTemplate, sanitizeTemplate } from './templateUtils';

export function downloadTemplate(template: Template, filename: string = 'template.json'): void {
  const jsonString = JSON.stringify(template, null, 2);
  const blob = new Blob([jsonString], { type: 'application/json' });
  const url = URL.createObjectURL(blob);
  
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  link.click();
  
  URL.revokeObjectURL(url);
}

export function downloadSeparateLayers(template: Template, baseName: string = 'template'): void {
  const layers = {
    ground: template.ground,
    static: template.static,
    monster: template.monster,
  };

  Object.entries(layers).forEach(([layerName, layerData]) => {
    const jsonString = JSON.stringify(layerData, null, 2);
    const blob = new Blob([jsonString], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    
    const link = document.createElement('a');
    link.href = url;
    link.download = `${baseName}_${layerName}.json`;
    link.click();
    
    URL.revokeObjectURL(url);
  });
}

export function copyToClipboard(template: Template): Promise<void> {
  const jsonString = JSON.stringify(template, null, 2);
  return navigator.clipboard.writeText(jsonString);
}

export function readFileAsText(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = (event) => {
      if (event.target?.result) {
        resolve(event.target.result as string);
      } else {
        reject(new Error('Failed to read file'));
      }
    };
    reader.onerror = () => reject(new Error('File reading error'));
    reader.readAsText(file);
  });
}

export async function importTemplate(file: File): Promise<{ template: Template; warnings: string[] }> {
  try {
    const content = await readFileAsText(file);
    const data = JSON.parse(content);
    
    const validation = validateTemplate(data);
    const warnings: string[] = [];
    
    if (!validation.isValid) {
      warnings.push('Template has validation errors. Attempting to sanitize...');
      try {
        const sanitized = sanitizeTemplate(data);
        warnings.push('Template has been automatically corrected.');
        return { template: sanitized, warnings };
      } catch (error) {
        throw new Error(`Cannot sanitize template: ${error}`);
      }
    }
    
    return { template: data as Template, warnings };
  } catch (error) {
    if (error instanceof SyntaxError) {
      throw new Error('Invalid JSON format');
    }
    throw error;
  }
}