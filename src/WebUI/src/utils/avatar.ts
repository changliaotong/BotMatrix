import { Bot, Users, Gamepad2, MessageSquare, Mail, Slack, Shield, Globe } from 'lucide-vue-next';

export const platforms = [
  { 
    id: 'Kook', 
    name: 'platform_kook', 
    icon: Gamepad2, 
    color: 'text-purple-500', 
  },
  { 
    id: 'DingTalk', 
    name: 'platform_dingtalk', 
    icon: MessageSquare, 
    color: 'text-blue-500', 
  },
  { 
    id: 'Email', 
    name: 'platform_email', 
    icon: Mail, 
    color: 'text-orange-500', 
  },
  { 
    id: 'Slack', 
    name: 'platform_slack', 
    icon: Slack, 
    color: 'text-red-500', 
  },
  { 
    id: 'TencentCloud', 
    name: 'platform_tencent', 
    icon: Shield, 
    color: 'text-indigo-500', 
  },
  { 
    id: 'WeCom', 
    name: 'platform_wecom', 
    icon: Globe, 
    color: 'text-blue-600', 
  },
  { 
    id: 'Web', 
    name: 'platform_web', 
    icon: Globe, 
    color: 'text-emerald-500', 
  },
  { 
    id: 'QQ', 
    name: 'platform_qq', 
    icon: MessageSquare, 
    color: 'text-blue-400', 
  }
];

export const getPlatformIcon = (platform: string) => {
  const p = platforms.find(p => p.id.toLowerCase() === platform.toLowerCase());
  return p ? p.icon : Bot;
};

export const getPlatformColor = (platform: string) => {
  const p = platforms.find(p => p.id.toLowerCase() === platform.toLowerCase());
  return p ? p.color : 'text-matrix';
};

export const isPlatformAvatar = (avatar: string | undefined | null) => {
  return avatar?.startsWith('platform://');
};

export const getPlatformFromAvatar = (avatar: string) => {
  return avatar.replace('platform://', '');
};

export const getAvatarUrl = (avatar: string | undefined | null) => {
  if (!avatar) return '';
  if (isPlatformAvatar(avatar)) return avatar;
  
  // 对于 QQ 头像或其他外部头像，使用后端代理以解决跨域/Referer 问题
  if (avatar.startsWith('http')) {
    // 检查是否已经是代理地址，避免重复代理
    if (avatar.includes('/api/proxy/avatar')) return avatar;
    return `/api/proxy/avatar?url=${encodeURIComponent(avatar)}`;
  }
  
  return avatar;
};
