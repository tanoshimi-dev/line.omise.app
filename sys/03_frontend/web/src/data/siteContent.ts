export const siteContent = {
  brand: 'はんなりdev LINEアプリ',
  nav: {
    demos: 'デモアプリ',
    features: '共通機能',
    whyUs: '選ばれる理由',
    contact: 'お問い合わせ',
  },
  hero: {
    badge: 'LINE公式アカウント連携',
    headline: 'はんなりdev\nLINEアプリ開発で\nビジネスを加速する',
    subtitle:
      'はんなりdevは、会員管理・予約システム・ECなど、LINE上で動くミニアプリを開発。\nお客様との接点を最大化し、業務効率を飛躍的に向上させます。',
    ctaPrimary: 'お問い合わせ',
    ctaSecondary: 'デモを見る',
  },
  demos: {
    sectionTitle: 'デモアプリ',
    sectionSubtitle:
      '実際に動作するデモアプリをご覧ください。すべてLINE上で完結するミニアプリとして開発しています。',
  },
  features: {
    sectionTitle: '共通技術基盤',
    sectionSubtitle:
      'すべてのアプリに共通する、信頼性の高い技術スタックと機能を採用しています。',
    items: [
      {
        title: 'LINE連携認証',
        description: 'LINEログインでワンタップ認証。ユーザー登録の手間を最小化します。',
        icon: 'shield',
      },
      {
        title: 'React + TypeScript',
        description: '型安全なフロントエンドで、高品質なUI/UXを実現します。',
        icon: 'code',
      },
      {
        title: 'Laravel API',
        description: '堅牢なバックエンドAPIで、安定したデータ処理を提供します。',
        icon: 'server',
      },
      {
        title: '管理者ダッシュボード',
        description: 'リアルタイムなデータ管理と分析を可能にする管理画面。',
        icon: 'dashboard',
      },
      {
        title: 'LINEプッシュ通知',
        description: '予約確認・ポイント付与など、重要な通知をLINEで即時配信。',
        icon: 'notification',
      },
      {
        title: 'レスポンシブデザイン',
        description: 'スマートフォン・タブレット・PCすべてに最適化されたUI。',
        icon: 'responsive',
      },
    ],
  },
  whyUs: {
    sectionTitle: '選ばれる理由',
    sectionSubtitle: '私たちが選ばれる3つの理由をご紹介します。',
    items: [
      {
        title: 'LINE上で完結するUX',
        description:
          'アプリのインストール不要。LINEの中でシームレスに動作するため、ユーザーの離脱を最小限に抑えます。',
        icon: 'smartphone',
      },
      {
        title: '実績ベースの開発力',
        description:
          '会員管理・予約・ECなど、多様な業種向けのアプリ開発実績。デモアプリで品質をご確認いただけます。',
        icon: 'trophy',
      },
      {
        title: '柔軟なカスタマイズ',
        description:
          'デモアプリをベースに、お客様のビジネスに最適な機能を追加・カスタマイズ。短期間での開発が可能です。',
        icon: 'customize',
      },
    ],
  },
  contact: {
    headline: 'LINEアプリ開発の\nご相談承ります',
    subtitle:
      'デモアプリの詳細説明、お見積もり、技術的なご質問など、\nLINE公式アカウントからお気軽にお問い合わせください。',
    cta: 'LINEで相談する',
  },
  footer: {
    copyright: `© ${new Date().getFullYear()} たのしみdev All rights reserved.`,
  },
  // LINE official account URL - update this with the real URL
  lineOaUrl: 'https://line.me/ti/p/3o83z6-mE0',
} as const
