import { message } from 'antd';
import type { ImportResult } from '../types';

export function maskKey(key: string): string {
  if (key.length <= 8) {
    return '*'.repeat(key.length);
  }
  return key.substring(0, 4) + '*'.repeat(key.length - 8) + key.substring(key.length - 4);
}

export function copyToClipboard(text: string): Promise<void> {
  return navigator.clipboard.writeText(text).then(() => {
    message.success('已复制到剪贴板');
  }).catch(() => {
    message.error('复制失败，请手动复制');
  });
}

export function getLimitText(source: { 
  billing_mode: 'count' | 'token'; 
  limit_count: number; 
  limit_tokens: number; 
  used_count: number; 
  used_tokens: number;
}): string {
  const limit = source.billing_mode === 'count' ? source.limit_count : source.limit_tokens;
  const used = source.billing_mode === 'count' ? source.used_count : source.used_tokens;
  return limit === 0 ? '无限制' : `${used}/${limit}`;
}

export function getStatus(source: { 
  is_active: boolean;
  billing_mode: 'count' | 'token';
  limit_count: number;
  limit_tokens: number;
  used_count: number;
  used_tokens: number;
}): { status: string; color: 'success' | 'warning' | 'error' } {
  if (!source.is_active) {
    return { status: '已禁用', color: 'error' };
  }
  const limit = source.billing_mode === 'count' ? source.limit_count : source.limit_tokens;
  const used = source.billing_mode === 'count' ? source.used_count : source.used_tokens;
  if (limit === 0 || used < limit) {
    return { status: '正常', color: 'success' };
  }
  return { status: '已用完', color: 'warning' };
}

export function renderImportResult(result: ImportResult): string {
  const lines: string[] = [];
  lines.push(`导入结果：`);
  lines.push(`新建对外模型: ${result.external_models_created}`);
  lines.push(`更新对外模型: ${result.external_models_updated}`);
  lines.push(`新建上游模型: ${result.sources_created}`);
  lines.push(`更新上游模型: ${result.sources_updated}`);
  
  if (result.errors?.length) {
    lines.push(`\n错误：`);
    result.errors.forEach(err => lines.push(`- ${err}`));
  }
  
  if (result.skipped?.length) {
    lines.push(`\n跳过：`);
    result.skipped.forEach(s => lines.push(`- ${s}`));
  }
  
  return lines.join('\n');
}

export function downloadBlob(blob: Blob, filename: string): void {
  const url = window.URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  window.URL.revokeObjectURL(url);
  document.body.removeChild(a);
}