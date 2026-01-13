export const newsContent: Record<string, Record<number, { title: string, summary: string, content: string, date: string, category: string }>> = {
  'zh-CN': {
    1: {
      title: '从机器人到数字员工：BotMatrix 的下一个十年',
      category: '愿景发布',
      date: '2025-12-01',
      summary: '我们宣布将重心全面转向“AI 原生智能体操作系统”，引入工号、职位及 KPI 考核等企业级人力资源管理概念。',
      content: `
        <p>从机器人到数字员工，再到自主进化的智能体蜂群，BotMatrix 的愿景始终是构建一个去中心化的 AI 未来。</p>
        <p>我们正在研发下一代 Swarm 编排引擎，它将支持成千上万个智能体在无人工干预的情况下，自主完成复杂的目标拆解与执行。</p>
        <h3>核心变革</h3>
        <ul>
          <li><strong>身份体系</strong>：每个智能体都拥有唯一的工号和职位。</li>
          <li><strong>KPI 考核</strong>：基于产出的自动化绩效评估系统。</li>
          <li><strong>组织架构</strong>：支持跨平台的智能体团队管理。</li>
        </ul>
      `
    },
    2: {
      title: 'Nexus Guard 引入分布式安全审计引擎',
      category: '架构演进',
      date: '2025-12-15',
      summary: '通过边缘节点并行审计技术，Nexus Guard 现在能够处理万人群聊中的瞬时消息洪峰，延迟低于 10ms。',
      content: `
        <p>安全是智能体协作的底线。Nexus Guard 分布式安全审计引擎的发布，为企业级应用提供了坚实的保障。</p>
        <p>该引擎采用边缘计算架构，将审计压力分摊到全球节点，确保在处理大规模并发消息时依然保持极低的延迟。</p>
        <h3>技术亮点</h3>
        <ul>
          <li><strong>毫秒级响应</strong>：平均处理延迟低于 10ms。</li>
          <li><strong>分布式架构</strong>：支持横向扩展以应对无限流量。</li>
          <li><strong>隐私保护</strong>：在审计过程中自动进行敏感数据脱敏。</li>
        </ul>
      `
    },
    3: {
      title: '早喵机器人“超级群管”插件 2.0 上线',
      category: '产品更新',
      date: '2025-12-28',
      summary: '新增积分经济联动系统与阶梯式违规禁言逻辑，让社群管理从“堵”转变为“疏”，大幅提升活跃度。',
      content: `
        <p>BotMatrix 的明星级助手“早喵”迎来了重大更新。2.0 版本不仅增强了管理能力，更引入了全新的激励机制。</p>
        <p>我们相信，优秀的社群不应仅仅依靠禁言，更应该通过正向的价值回馈来驱动成员参与。</p>
        <h3>新功能预览</h3>
        <ul>
          <li><strong>积分商城</strong>：成员可以通过贡献换取特定的机器人服务。</li>
          <li><strong>智慧禁言</strong>：根据违规历史自动调整禁言时长。</li>
          <li><strong>AI 话题引导</strong>：自动识别社群热点并生成互动话题。</li>
        </ul>
      `
    },
    4: {
      title: 'Global Agent Mesh 协议正式发布：打破企业间 AI 孤岛',
      category: '技术突破',
      date: '2026-01-04',
      summary: '我们推出了基于去中心化握手协议的 Agent 协作网络，支持跨企业的数字员工身份委派与技能发现。',
      content: `
        <p>今天，我们非常自豪地宣布 Global Agent Mesh 协议正式发布。这是一个里程碑式的时刻，标志着分布式智能体协作进入了一个全新的阶段。</p>
        <h3>核心特性</h3>
        <ul>
          <li><strong>去中心化握手</strong>：无需中心服务器即可建立安全连接。</li>
          <li><strong>身份委派</strong>：支持跨企业的智能体身份认证。</li>
          <li><strong>技能发现</strong>：自动匹配并调用网络中的专业智能体。</li>
        </ul>
        <p>我们相信，Global Agent Mesh 将成为未来 AI 原生操作系统的基石。欢迎访问我们的 GitHub 仓库参与贡献。</p>
      `
    }
  },
  'en-US': {
    1: {
      title: 'From Robots to Digital Employees: The Next Decade of BotMatrix',
      category: 'Vision Release',
      date: '2025-12-01',
      summary: 'We announced a strategic shift towards "AI Native Agent OS", introducing enterprise concepts like employee IDs, positions, and KPI assessments.',
      content: `
        <p>From bots to digital employees, and then to autonomously evolving agent swarms, BotMatrix\'s vision has always been to build a decentralized AI future.</p>
        <p>We are developing the next generation Swarm orchestration engine, which will support thousands of agents in autonomously completing complex goal decomposition and execution without human intervention.</p>
        <h3>Core Changes</h3>
        <ul>
          <li><strong>Identity System</strong>: Every agent has a unique employee ID and position.</li>
          <li><strong>KPI Assessment</strong>: Automated performance evaluation system based on output.</li>
          <li><strong>Organizational Structure</strong>: Supports cross-platform agent team management.</li>
        </ul>
      `
    },
    2: {
      title: 'Nexus Guard Introduces Distributed Security Audit Engine',
      category: 'Arch Evolution',
      date: '2025-12-15',
      summary: 'Through edge node parallel auditing, Nexus Guard can now handle instantaneous message peaks in groups of 10,000+, with latency below 10ms.',
      content: `
        <p>Security is the bottom line of agent collaboration. The release of the Nexus Guard distributed security audit engine provides a solid guarantee for enterprise-level applications.</p>
        <p>The engine adopts an edge computing architecture, distributing audit pressure to global nodes to ensure extremely low latency while processing large-scale concurrent messages.</p>
        <h3>Technical Highlights</h3>
        <ul>
          <li><strong>Millisecond Response</strong>: Average processing latency below 10ms.</li>
          <li><strong>Distributed Architecture</strong>: Supports horizontal scaling to handle infinite traffic.</li>
          <li><strong>Privacy Protection</strong>: Automatically desensitizes sensitive data during the audit process.</li>
        </ul>
      `
    },
    3: {
      title: 'ZaoMiao Robot "Super Group Admin" Plugin 2.0 Online',
      category: 'Product Update',
      date: '2025-12-28',
      summary: 'Added integration with points economy and tiered violation mute logic, transforming community management from "blocking" to "guiding".',
      content: `
        <p>BotMatrix\'s star assistant "ZaoMiao" has received a major update. Version 2.0 not only enhances management capabilities but also introduces a brand new incentive mechanism.</p>
        <p>We believe that an excellent community should not rely solely on muting, but should drive member participation through positive value feedback.</p>
        <h3>New Features Preview</h3>
        <ul>
          <li><strong>Points Mall</strong>: Members can exchange contributions for specific robot services.</li>
          <li><strong>Smart Mute</strong>: Automatically adjusts mute duration based on violation history.</li>
          <li><strong>AI Topic Guidance</strong>: Automatically identifies community hotspots and generates interactive topics.</li>
        </ul>
      `
    },
    4: {
      title: 'Global Agent Mesh Protocol Officially Released: Breaking Enterprise AI Silos',
      category: 'Tech Breakthrough',
      date: '2026-01-04',
      summary: 'We introduced an Agent collaboration network based on decentralized handshake protocols, supporting cross-enterprise identity delegation and skill discovery.',
      content: `
        <p>Today, we are proud to announce the official release of the Global Agent Mesh protocol. This is a milestone moment marking a new era of distributed agent collaboration.</p>
        <h3>Key Features</h3>
        <ul>
          <li><strong>Decentralized Handshake</strong>: Establishes secure connections without a central server.</li>
          <li><strong>Identity Delegation</strong>: Supports agent identity authentication across enterprises.</li>
          <li><strong>Skill Discovery</strong>: Automatically matches and invokes professional agents in the network.</li>
        </ul>
        <p>We believe Global Agent Mesh will become the cornerstone of future AI-native operating systems. Welcome to our GitHub repository to contribute.</p>
      `
    }
  },
  'zh-TW': {
    1: {
      title: '從機器人到數位員工：BotMatrix 的下一個十年',
      category: '願景發布',
      date: '2025-12-01',
      summary: '我們宣布將重心全面轉向「AI 原生智慧體作業系統」，引入工號、職位及 KPI 考核等企業級人力資源管理概念。',
      content: `
        <p>從機器人到數位員工，再到自主進化的智慧體蜂群，BotMatrix 的願景始終是構建一個去中心化的 AI 未來。</p>
        <p>我們正在研發下一代 Swarm 編排引擎，它將支持成千上萬個智慧體在無人工干預的情況下，自主完成複雜的目標拆解與執行。</p>
        <h3>核心變革</h3>
        <ul>
          <li><strong>身份體系</strong>：每個智慧體都擁有唯一的工號和職位。</li>
          <li><strong>KPI 考核</strong>：基於產出的自動化績效評估系統。</li>
          <li><strong>組織架構</strong>：支持跨平台的智慧體團隊管理。</li>
        </ul>
      `
    },
    2: {
      title: 'Nexus Guard 引入分佈式安全審計引擎',
      category: '架構演進',
      date: '2025-12-15',
      summary: '透過邊緣節點并行審計技術，Nexus Guard 現在能夠處理萬人群聊中的瞬時消息洪峰，延遲低於 10ms。',
      content: `
        <p>安全是智慧體協作的底線。Nexus Guard 分佈式安全審計引擎的發布，為企業級應用提供了堅實保障。</p>
        <p>該引擎採用邊緣計算架構，將審計壓力分攤到全球節點，確保在處理大規模併發消息時依然保持極低的延遲。</p>
        <h3>技術亮點</h3>
        <ul>
          <li><strong>毫秒級響應</strong>：平均處理延遲低於 10ms。</li>
          <li><strong>分佈式架構</strong>：支持橫向擴展以應對無限流量。</li>
          <li><strong>隱私保護</strong>：在審計過程中自動進行敏感數據脫敏。</li>
        </ul>
      `
    },
    3: {
      title: '早喵機器人「超級群管」插件 2.0 上線',
      category: '產品更新',
      date: '2025-12-28',
      summary: '新增積分經濟聯動系統與階梯式違規禁言邏輯，讓社群管理從「堵」轉變為「疏」，大幅提升活躍度。',
      content: `
        <p>BotMatrix 的明星級助手「早喵」迎來了重大更新。2.0 版本不僅增強了管理能力，更引入了全新的激勵機制。</p>
        <p>我們相信，優秀的社群不應僅僅依靠禁言，更應該通過正向的價值回饋來驅動成員參與。</p>
        <h3>新功能預覽</h3>
        <ul>
          <li><strong>積分商城</strong>：成員可以通過貢獻換取特定的機器人服務。</li>
          <li><strong>智慧禁言</strong>：根據違規歷史自動調整禁言時長。</li>
          <li><strong>AI 話題引導</strong>：自動識別社群熱點並生成互動話題。</li>
        </ul>
      `
    },
    4: {
      title: 'Global Agent Mesh 協議正式發布：打破企業間 AI 孤島',
      category: '技術突破',
      date: '2026-01-04',
      summary: '我們推出了基於去中心化握手協議的 Agent 協作網路，支援跨企業的數位員工身份委派與技能發現。',
      content: `
        <p>今天，我們非常自豪地宣布 Global Agent Mesh 協議正式發布。這是一個里程碑式的時刻，標誌著分佈式智慧體協作進入了一個全新的階段。</p>
        <h3>核心特性</h3>
        <ul>
          <li><strong>去中心化握手</strong>：無需中心伺服器即可建立安全連接。</li>
          <li><strong>身份委派</strong>：支持跨企業的智慧體身份認證。</li>
          <li><strong>技能發現</strong>：自動匹配並調用網路中的專業智慧體。</li>
        </ul>
        <p>我們相信，Global Agent Mesh 將成為未來 AI 原生作業系統的基石。歡迎訪問我們的 GitHub 倉庫參與貢獻。</p>
      `
    }
  },
  'ja-JP': {
    1: {
      title: 'ロボットからデジタル従業員へ：BotMatrix の次の 10 年',
      category: 'ビジョン発表',
      date: '2025-12-01',
      summary: '「AI ネイティブ Agent OS」への戦略的転換を発表し、従業員 ID、役職、KPI 評価などの企業レベルの概念を導入しました。',
      content: `
        <p>ロボットからデジタル従業員へ、そして自律的に進化するエージェントスウォームへ。BotMatrix のビジョンは常に、分散型の AI の未来を築くことにあります。</p>
        <p>私たちは次世代の Swarm オーケストレーションエンジンを開発しています。これにより、何千ものエージェントが人間の介入なしに、複雑な目標の分解と実行を自律的に完了できるようになります。</p>
        <h3>核心的な変革</h3>
        <ul>
          <li><strong>アイデンティティ体系</strong>：各エージェントは一意の従業員 ID と役職を持ちます。</li>
          <li><strong>KPI 評価</strong>：成果に基づく自動パフォーマンス評価システム。</li>
          <li><strong>組織構造</strong>：プラットフォームを跨いだエージェントチームの管理をサポート。</li>
        </ul>
      `
    },
    2: {
      title: 'Nexus Guard が分散型セキュリティ監査エンジンを導入',
      category: 'アーキテクチャの進化',
      date: '2025-12-15',
      summary: 'エッジノード並列監査技術により、Nexus Guard は 1 万人以上のグループでの瞬間的なメッセージピークを 10ms 未満の遅延で処理できるようになりました。',
      content: `
        <p>セキュリティはエージェント連携の生命線です。Nexus Guard 分散型セキュリティ監査エンジンのリリースは、企業レベルのアプリケーションに強固な保証を提供します。</p>
        <p>このエンジンはエッジコンピューティングアーキテクチャを採用し、監査の負荷をグローバルノードに分散させることで、大規模な同時実行メッセージの処理中も極めて低い遅延を維持します。</p>
        <h3>技術的なハイライト</h3>
        <ul>
          <li><strong>ミリ秒単位のレスポンス</strong>：平均処理遅延は 10ms 未満。</li>
          <li><strong>分散型アーキテクチャ</strong>：無限のトラフィックに対応するための水平スケーリングをサポート。</li>
          <li><strong>プライバシー保護</strong>：監査プロセス中に機密データを自動的に匿名化。</li>
        </ul>
      `
    },
    3: {
      title: '早喵ロボット「スーパー群管理」プラグイン 2.0 公開',
      category: '製品アップデート',
      date: '2025-12-28',
      summary: 'ポイント経済との連動システムと段階的な違反ミュートロジックを追加し、コミュニティ管理を「ブロック」から「誘導」へと変革しました。',
      content: `
        <p>BotMatrix のスターアシスタント「早喵（ザオミャオ）」が大幅なアップデートを迎えました。バージョン 2.0 では管理能力が強化されただけでなく、全く新しいインセンティブメカニズムが導入されました。</p>
        <p>私たちは、優れたコミュニティは単なるミュート（禁言）に頼るのではなく、ポジティブな価値のフィードバックを通じてメンバーの参加を促すべきだと信じています。</p>
        <h3>新機能プレビュー</h3>
        <ul>
          <li><strong>ポイントモール</strong>：メンバーは貢献を通じて特定のロボットサービスと交換できます。</li>
          <li><strong>インテリジェントミュート</strong>：違反履歴に基づいてミュート時間を自動的に調整します。</li>
          <li><strong>AI トピックガイダンス</strong>：コミュニティのホットスポットを自動的に識別し、交流トピックを生成します。</li>
        </ul>
      `
    },
    4: {
      title: 'Global Agent Mesh プロトコル正式リリース：企業間 AI 孤島を打破',
      category: '技術的突破',
      date: '2026-01-04',
      summary: '分散型ハンドシェイクプロトコルに基づく Agent 連携ネットワークを導入し、企業を跨いだデジタル従業員のアイデンティティ委任とスキル発見をサポートします。',
      content: `
        <p>本日、Global Agent Mesh プロトコルが正式にリリースされたことを誇りを持って発表します。これは、分散型エージェント連携が全く新しい段階に入ったことを示す記念すべき瞬間です。</p>
        <h3>主な特徴</h3>
        <ul>
          <li><strong>分散型ハンドシェイク</strong>：中央サーバーなしで安全な接続を確立。</li>
          <li><strong>アイデンティティ委任</strong>：企業を跨いだエージェントのアイデンティティ認証をサポート。</li>
          <li><strong>スキル発見</strong>：ネットワーク内の専門エージェントを自動的にマッチングし呼び出します。</li>
        </ul>
        <p>Global Agent Mesh は将来の AI ネイティブ OS の礎になると信じています。GitHub リポジトリへの貢献を歓迎します。</p>
      `
    }
  }
};

