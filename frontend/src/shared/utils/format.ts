// frontend/src/shared/utils/format.ts

/**
 * 格式化文件大小为人类可读字符串
 * @param bytes 字节数
 */
export function formatFileSize(bytes: number): string {
    if (bytes < 1024) return bytes + ' B'
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
    if (bytes < 1024 * 1024 * 1024) return (bytes / 1024 / 1024).toFixed(1) + ' MB'
    return (bytes / 1024 / 1024 / 1024).toFixed(2) + ' GB'
}

/**
 * 格式化传输速度
 * @param bytesPerSec 每秒字节数
 */
export function formatSpeed(bytesPerSec: number): string {
    return formatFileSize(bytesPerSec) + '/s'
}
