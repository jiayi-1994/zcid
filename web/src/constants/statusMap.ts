export interface StatusStyle {
  color: string;
  bg: string;
  icon: string;
  label: string;
}

export const STATUS_MAP = {
  success: {
    color: '#00B42A',
    bg: '#E8FFEA',
    icon: 'IconCheckCircle',
    label: '成功',
  },
  running: {
    color: '#1677FF',
    bg: '#E8F3FF',
    icon: 'IconLoading',
    label: '运行中',
  },
  failed: {
    color: '#F53F3F',
    bg: '#FFECE8',
    icon: 'IconCloseCircle',
    label: '失败',
  },
  warning: {
    color: '#FF7D00',
    bg: '#FFF7E8',
    icon: 'IconExclamation',
    label: '警告',
  },
  pending: {
    color: '#86909C',
    bg: '#F2F3F5',
    icon: 'IconClockCircle',
    label: '等待中',
  },
  cancelled: {
    color: '#86909C',
    bg: '#F2F3F5',
    icon: 'IconMinusCircle',
    label: '已取消',
  },
  timeout: {
    color: '#F53F3F',
    bg: '#FFECE8',
    icon: 'IconClockCircle',
    label: '超时',
  },
} as const satisfies Record<string, StatusStyle>;
