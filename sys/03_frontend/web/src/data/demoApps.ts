export interface DemoApp {
  id: string
  name: string
  description: string
  accentColor: string
  accentColorLight: string
  features: string[]
  techBadges: string[]
  screenshot?: string
  movie?: string
}

export const demoApps: DemoApp[] = [
  {
    id: 'membership',
    name: '会員管理アプリ',
    description:
      'デジタル会員証の発行からポイント管理、ランク制度まで。店舗の会員管理をLINEで完結させるアプリです。',
    accentColor: '#6366f1',
    accentColorLight: '#eef2ff',
    features: [
      'デジタル会員証（QRコード）',
      'ポイント管理・履歴',
      '会員ランク制度',
      '管理者ダッシュボード',
    ],
    techBadges: ['React', 'TypeScript', 'Laravel', 'LINE LIFF'],
    screenshot: '/images/membership.png',
    movie: '/images/membership.mp4',
  },
  {
    id: 'salon-reservation',
    name: 'サロン予約アプリ',
    description:
      'サービス選択からスタッフ指名、日時指定まで4ステップで予約完了。サロン業務を効率化するLINE予約アプリです。',
    accentColor: '#ec4899',
    accentColorLight: '#fdf2f8',
    features: [
      '4ステップ簡単予約',
      'スタッフスケジュール管理',
      'チャット機能',
      'LINE通知連携',
    ],
    techBadges: ['React', 'TypeScript', 'Laravel', 'LINE Messaging API'],
    screenshot: '/images/salon-reservation.png',
    movie: '/images/salon-reservation.mp4',
  },
  {
    id: 'sweets-shop',
    name: 'スイーツショップアプリ',
    description:
      '商品ギャラリーからポイント付与、レビュー投稿まで。スイーツ店の顧客体験を向上させるLINEアプリです。',
    accentColor: '#f59e0b',
    accentColorLight: '#fffbeb',
    features: [
      '商品ギャラリー・カテゴリ',
      'ポイントQRコード',
      'レビュー・口コミ',
      'ニュース配信',
    ],
    techBadges: ['React', 'TypeScript', 'Laravel', 'LINE Messaging API'],
    screenshot: '/images/sweets-shop.png',
    movie: '/images/sweets-shop.mp4',
  },
]
