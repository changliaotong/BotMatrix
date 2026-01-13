export interface DocItem {
  title: string;
  category: string;
  date: string;
  summary: string;
  content: string;
}

export const docsContent: Record<string, Record<number, DocItem>> = {
  'zh-CN': {
    1: {
      title: 'Global Agent Mesh 架构指南',
      category: '核心架构',
      date: '2026-01-10',
      summary: '深入了解 BotMatrix 的分布式 Agent 协作网络，探索跨域发现与 B2B 协作逻辑。',
      content: `
        <h3>系统架构概览</h3>
        <p>BotMatrix 是一个采用分布式、解耦设计的机器人矩阵管理系统。它通过核心的消息分发中心与多个执行节点协作，实现了高并发和高可扩展性。</p>
        <h4>核心组件</h4>
        <ul>
          <li><strong>BotNexus (中心控制节点)</strong>：系统的“大脑”和“路由器”，处理 AI 意图识别、Agent Mesh 枢纽及 MCP Host 管理。</li>
          <li><strong>BotWorker (任务执行节点)</strong>：实际处理业务逻辑的“四肢”，负责 AI 推理、MCP 工具执行及 RAG 2.0 检索。</li>
          <li><strong>Redis 通信总线</strong>：利用 Pub/Sub 机制实现节点间的实时通信与任务分发。</li>
        </ul>
        <h3>Global Agent Mesh 特性</h3>
        <p>作为 Mesh 网络的枢纽，BotNexus 处理跨域发现、联邦身份验证与 B2B 协作逻辑，确保不同企业间的 Agent 能够安全、高效地进行能力互补。</p>
      `
    },
    2: {
      title: '智能体集群 Swarm 引擎',
      category: '进阶技术',
      date: '2026-01-11',
      summary: '学习如何利用 Swarm 引擎实现多智能体协作与动态任务分解。',
      content: `
        <h3>什么是 Swarm 引擎？</h3>
        <p>Swarm 引擎是 BotMatrix 用于处理复杂任务的协作框架。它能够将一个宏大目标拆解为多个子任务，并分配给最适合的专业 Agent 执行。</p>
        <h4>关键机制</h4>
        <ul>
          <li><strong>任务分解器</strong>：利用 LLM 将复杂指令转化为有序的任务流。</li>
          <li><strong>动态接力</strong>：Agent 之间可以根据处理结果动态转移任务控制权。</li>
          <li><strong>集体智能</strong>：通过多 Agent 投票或共识机制提升决策准确性。</li>
        </ul>
        <p>在 2026 Q3 的规划中，Swarm 引擎将支持更加复杂的任务编排与分布式共识算法。</p>
      `
    },
    3: {
      title: '计算机使用 (Computer Use) 指南',
      category: '前沿探索',
      date: '2026-01-12',
      summary: '探索 Agent 如何像人类一样操作桌面环境与 Web 应用。',
      content: `
        <h3>让 Agent 拥有“双手”</h3>
        <p>Computer Use 技术允许 BotMatrix 的 Agent 直接与操作系统交互，包括点击、拖拽、输入及视觉解析。</p>
        <h4>应用场景</h4>
        <ul>
          <li><strong>自动化测试</strong>：自动完成跨平台的 UI 回归测试。</li>
          <li><strong>数据采集</strong>：在没有 API 的老旧系统中提取结构化数据。</li>
          <li><strong>流程自动化</strong>：模拟人类在多个软件间流转业务流程。</li>
        </ul>
        <p>该功能目前集成在 BotWorker 的扩展模块中，支持 Linux 及 Windows 的无头模式运行。</p>
      `
    },
    4: {
      title: '数字化员工系统设定',
      category: '产品方案',
      date: '2026-01-13',
      summary: '如何构建具备认知、身份与 KPI 的独立生产力单元。',
      content: `
        <h3>数字化员工的概念</h3>
        <p>不仅仅是对话机器人，数字化员工是具备完整身份标识、权限体系及业务能力的 Agent。它们在系统中被赋予唯一的 IdentityGORM 标识。</p>
        <h4>核心能力</h4>
        <ul>
          <li><strong>意图调度器</strong>：精准匹配用户需求与员工技能。</li>
          <li><strong>操作审计</strong>：全量记录 Agent 的操作路径，确保合规与可追溯。</li>
          <li><strong>技能集 (Toolset)</strong>：通过 MCP 协议动态加载的业务能力包。</li>
        </ul>
      `
    },
    5: {
      title: '认知记忆与 RAG 2.0',
      category: '人工智能',
      date: '2026-01-14',
      summary: '基于向量数据库与知识图谱的持久化记忆方案。',
      content: `
        <h3>从检索到认知</h3>
        <p>BotMatrix 的 RAG 2.0 不仅仅是简单的向量检索，它引入了 Agentic RAG 与 GraphRAG 技术。</p>
        <h4>技术要点</h4>
        <ul>
          <li><strong>短期记忆</strong>：基于 Redis 的滑动窗口上下文管理。</li>
          <li><strong>长期记忆</strong>：基于向量数据库的语义检索。</li>
          <li><strong>知识图谱</strong>：利用 Neo4j 等图数据库处理跨文档的复杂关系推理。</li>
        </ul>
      `
    },
    6: {
      title: 'MCP 插件开发手册',
      category: '开发者中心',
      date: '2026-01-15',
      summary: '遵循 Model Context Protocol 标准，快速扩展 Agent 能力。',
      content: `
        <h3>MCP 标准协议</h3>
        <p>Model Context Protocol (MCP) 是 BotMatrix 推荐的插件开发标准，实现了能力提供方与消费方的解耦。</p>
        <h4>开发步骤</h4>
        <ul>
          <li><strong>定义资源 (Resources)</strong>：暴露静态数据或动态文档。</li>
          <li><strong>暴露工具 (Tools)</strong>：提供可被 LLM 调用的函数接口。</li>
          <li><strong>预设提示词 (Prompts)</strong>：封装常用的业务提示词模板。</li>
        </ul>
        <p>支持 stdio, SSE 及 WebSocket 多种传输方式，兼容 Python, Go, Node.js 等主流语言。</p>
      `
    }
  },
  'en-US': {
    1: {
      title: 'Global Agent Mesh Architecture Guide',
      category: 'Core Architecture',
      date: '2026-01-10',
      summary: 'Deep dive into BotMatrix\'s distributed Agent collaboration network, exploring cross-domain discovery and B2B collaboration logic.',
      content: `
        <h3>System Overview</h3>
        <p>BotMatrix is a distributed, decoupled chatbot management system. It achieves high concurrency and scalability through a central message distribution hub and multiple execution nodes.</p>
        <h4>Core Components</h4>
        <ul>
          <li><strong>BotNexus (Central Control Node)</strong>: The "brain" and "router" of the system, handling AI intent recognition, Agent Mesh hub, and MCP Host management.</li>
          <li><strong>BotWorker (Execution Node)</strong>: The "limbs" handling business logic, responsible for AI inference, MCP tool execution, and RAG 2.0 retrieval.</li>
          <li><strong>Redis Communication Bus</strong>: Uses Pub/Sub mechanism for real-time communication and task distribution between nodes.</li>
        </ul>
        <h3>Global Agent Mesh Features</h3>
        <p>As the hub of the Mesh network, BotNexus handles cross-domain discovery, federated authentication, and B2B collaboration logic, ensuring secure and efficient capability complementarity between Agents from different enterprises.</p>
      `
    },
    2: {
      title: 'Intelligent Swarm Engine',
      category: 'Advanced Tech',
      date: '2026-01-11',
      summary: 'Learn how to use Swarm engine for multi-agent collaboration and dynamic task decomposition.',
      content: `
        <h3>What is Swarm Engine?</h3>
        <p>The Swarm engine is BotMatrix\'s collaboration framework for complex tasks. It decomposes grand goals into multiple sub-tasks assigned to the most suitable specialized Agents.</p>
        <h4>Key Mechanisms</h4>
        <ul>
          <li><strong>Task Decomposer</strong>: Uses LLM to transform complex instructions into ordered task flows.</li>
          <li><strong>Dynamic Handoff</strong>: Agents can dynamically transfer task control based on processing results.</li>
          <li><strong>Collective Intelligence</strong>: Enhances decision accuracy through multi-agent voting or consensus mechanisms.</li>
        </ul>
        <p>In the 2026 Q3 roadmap, the Swarm engine will support even more complex task orchestration and distributed consensus algorithms.</p>
      `
    },
    3: {
      title: 'Computer Use Guide',
      category: 'Cutting-edge',
      date: '2026-01-12',
      summary: 'Explore how Agents operate desktop environments and web apps like humans.',
      content: `
        <h3>Giving Agents "Hands"</h3>
        <p>Computer Use technology allows BotMatrix Agents to interact directly with the operating system, including clicking, dragging, typing, and visual parsing.</p>
        <h4>Application Scenarios</h4>
        <ul>
          <li><strong>Automated Testing</strong>: Automatically complete cross-platform UI regression testing.</li>
          <li><strong>Data Collection</strong>: Extract structured data from legacy systems without APIs.</li>
          <li><strong>Process Automation</strong>: Simulate human workflow across multiple software applications.</li>
        </ul>
        <p>This feature is currently integrated into the BotWorker expansion module, supporting headless mode on Linux and Windows.</p>
      `
    },
    4: {
      title: 'Digital Employee System',
      category: 'Product Solution',
      date: '2026-01-13',
      summary: 'Building independent productivity units with cognition, identity, and KPIs.',
      content: `
        <h3>Concept of Digital Employee</h3>
        <p>More than just chatbots, Digital Employees are Agents with full identity, permissions, and business capabilities, identified by unique IdentityGORM in the system.</p>
        <h4>Core Capabilities</h4>
        <ul>
          <li><strong>Intent Scheduler</strong>: Precisely matches user needs with employee skills.</li>
          <li><strong>Operation Audit</strong>: Full logging of Agent operation paths to ensure compliance and traceability.</li>
          <li><strong>Toolset</strong>: Business capability packages dynamically loaded via MCP protocol.</li>
        </ul>
      `
    },
    5: {
      title: 'Cognitive Memory & RAG 2.0',
      category: 'AI',
      date: '2026-01-14',
      summary: 'Persistent memory solutions based on vector databases and knowledge graphs.',
      content: `
        <h3>From Retrieval to Cognition</h3>
        <p>BotMatrix RAG 2.0 introduces Agentic RAG and GraphRAG technologies beyond simple vector retrieval.</p>
        <h4>Technical Highlights</h4>
        <ul>
          <li><strong>Short-term Memory</strong>: Sliding window context management based on Redis.</li>
          <li><strong>Long-term Memory</strong>: Semantic retrieval based on vector databases.</li>
          <li><strong>Knowledge Graph</strong>: Uses graph databases like Neo4j for complex relationship reasoning across documents.</li>
        </ul>
      `
    },
    6: {
      title: 'MCP Plugin Development Manual',
      category: 'Dev Center',
      date: '2026-01-15',
      summary: 'Extend Agent capabilities quickly following the Model Context Protocol standard.',
      content: `
        <h3>MCP Protocol Standard</h3>
        <p>Model Context Protocol (MCP) is the recommended standard for BotMatrix plugins, decoupling capability providers and consumers.</p>
        <h4>Development Steps</h4>
        <ul>
          <li><strong>Define Resources</strong>: Expose static data or dynamic documents.</li>
          <li><strong>Expose Tools</strong>: Provide function interfaces callable by LLMs.</li>
          <li><strong>Preset Prompts</strong>: Encapsulate common business prompt templates.</li>
        </ul>
        <p>Supports stdio, SSE, and WebSocket transport methods, compatible with Python, Go, Node.js, and other mainstream languages.</p>
      `
    }
  },
  'zh-TW': {
    1: {
      title: 'Global Agent Mesh 架構指南',
      category: '核心架構',
      date: '2026-01-10',
      summary: '深入了解 BotMatrix 的分佈式 Agent 協作網絡，探索跨域發現與 B2B 協作邏輯。',
      content: `
        <h3>系統架構概覽</h3>
        <p>BotMatrix 是一個採用分佈式、解耦設計的機器人矩陣管理系統。它通過核心的消息分發中心與多個執行節點協作，實現了高併發和高可擴展性。</p>
        <h4>核心組件</h4>
        <ul>
          <li><strong>BotNexus (中心控制節點)</strong>：系統的“大腦”和“路由器”，處理 AI 意圖識別、Agent Mesh 樞紐及 MCP Host 管理。</li>
          <li><strong>BotWorker (任務執行節點)</strong>：實際處理業務邏輯的“四肢”，負責 AI 推理、MCP 工具執行及 RAG 2.0 檢索。</li>
          <li><strong>Redis 通信總線</strong>：利用 Pub/Sub 機制實現節點間的實時通信與任務分發。</li>
        </ul>
        <h3>Global Agent Mesh 特性</h3>
        <p>作為 Mesh 網絡的樞紐，BotNexus 處理跨域發現、聯邦身份驗證與 B2B 協作邏輯，確保不同企業間的 Agent 能夠安全、高效地進行能力互補。</p>
      `
    },
    2: {
      title: '智慧體集群 Swarm 引擎',
      category: '進階技術',
      date: '2026-01-11',
      summary: '學習如何利用 Swarm 引擎實現多智慧體協作與動態任務分解。',
      content: `
        <h3>什麼是 Swarm 引擎？</h3>
        <p>Swarm 引擎是 BotMatrix 用於處理複雜任務的協作框架。它能夠將一個宏大目標拆解為多個子任務，並分配給最適合的專業 Agent 執行。</p>
        <h4>關鍵機制</h4>
        <ul>
          <li><strong>任務分解器</strong>：利用 LLM 將複雜指令轉化為有序的任務流。</li>
          <li><strong>動態接力</strong>：Agent 之間可以根據處理結果動態轉移任務控制權。</li>
          <li><strong>集體智能</strong>：通過多 Agent 投票或共識機制提升決策準確性。</li>
        </ul>
        <p>在 2026 Q3 的規劃中，Swarm 引擎將支持更加複雜的任务編排與分佈式共識算法。</p>
      `
    },
    3: {
      title: '電腦使用 (Computer Use) 指南',
      category: '前沿探索',
      date: '2026-01-12',
      summary: '探索 Agent 如何像人類一樣操作桌面環境與 Web 應用。',
      content: `
        <h3>讓 Agent 擁有“雙手”</h3>
        <p>Computer Use 技術允許 BotMatrix 的 Agent 直接與操作系統交互，包括點擊、拖拽、輸入及視覺解析。</p>
        <h4>應用場景</h4>
        <ul>
          <li><strong>自動化測試</strong>：自動完成跨平台的 UI 回歸測試。</li>
          <li><strong>數據采集</strong>：在沒有 API 的老舊系統中提取結構化數據。</li>
          <li><strong>流程自動化</strong>：模擬人類在多個軟件間流轉業務流程。</li>
        </ul>
        <p>該功能目前集成在 BotWorker 的擴展模塊中，支持 Linux 及 Windows 的無頭模式運行。</p>
      `
    },
    4: {
      title: '數位員工系統設定',
      category: '產品方案',
      date: '2026-01-13',
      summary: '如何構建具備認知、身份與 KPI 的獨立生產力單元。',
      content: `
        <h3>數位員工的概念</h3>
        <p>不僅僅是對話機器人，數位員工是具備完整身份標識、權限體系及業務能力的 Agent。它們在系統中被賦予唯一的 IdentityGORM 標識。</p>
        <h4>核心能力</h4>
        <ul>
          <li><strong>意圖調度器</strong>：精準匹配用戶需求與員工技能。</li>
          <li><strong>操作審計</strong>：全量記錄 Agent 的操作路徑，確保合規與可追蹤。</li>
          <li><strong>技能集 (Toolset)</strong>：通過 MCP 協議動態加載的業務能力包。</li>
        </ul>
      `
    },
    5: {
      title: '認知記憶與 RAG 2.0',
      category: '人工智能',
      date: '2026-01-14',
      summary: '基於向量數據庫與知識圖譜的持久化記憶方案。',
      content: `
        <h3>從檢索到認知</h3>
        <p>BotMatrix 的 RAG 2.0 不僅僅是簡單的向量檢索，它引入了 Agentic RAG 與 GraphRAG 技術。</p>
        <h4>技術要點</h4>
        <ul>
          <li><strong>短期記憶</strong>：基於 Redis 的滑動窗口上下文管理。</li>
          <li><strong>長期記憶</strong>：基於向量數據庫的語義檢索。</li>
          <li><strong>知識圖譜</strong>：利用 Neo4j 等圖數據庫處理跨文檔的複雜關係推理。</li>
        </ul>
      `
    },
    6: {
      title: 'MCP 插件開發手冊',
      category: '開發者中心',
      date: '2026-01-15',
      summary: '遵循 Model Context Protocol 標準，快速擴展 Agent 能力。',
      content: `
        <h3>MCP 標準協議</h3>
        <p>Model Context Protocol (MCP) 是 BotMatrix 推薦的插件開發標準，實現了能力提供方與消費方的解耦。</p>
        <h4>開發步驟</h4>
        <ul>
          <li><strong>定義資源 (Resources)</strong>：暴露靜態數據或動態文檔。</li>
          <li><strong>暴露工具 (Tools)</strong>：提供可被 LLM 調用的函數接口。</li>
          <li><strong>預設提示詞 (Prompts)</strong>：封裝常用的業務提示詞模板。</li>
        </ul>
        <p>支持 stdio, SSE 及 WebSocket 多種傳輸方式，兼容 Python, Go, Node.js 等主流語言。</p>
      `
    }
  },
  'ja-JP': {
    1: {
      title: 'Global Agent Mesh アーキテクチャガイド',
      category: 'コアアーキテクチャ',
      date: '2026-01-10',
      summary: 'BotMatrix の分散型エージェント連携ネットワークを深く掘り下げ、ドメインを越えた発見と B2B 連携ロジックを探索します。',
      content: `
        <h3>システムアーキテクチャの概要</h3>
        <p>BotMatrix は、分散型で疎結合な設計を採用したボットマトリックス管理システムです。コアとなるメッセージ配信センターと複数の実行ノードが連携することで、高い並行性と拡張性を実現しています。</p>
        <h4>コアコンポーネント</h4>
        <ul>
          <li><strong>BotNexus (中央制御ノード)</strong>：システムの「脳」であり「ルーター」です。AI 意図認識、Agent Mesh ハブ、MCP ホスト管理を処理します。</li>
          <li><strong>BotWorker (タスク実行ノード)</strong>：ビジネスロジックを実際に処理する「手足」であり、AI 推論、MCP ツール実行、RAG 2.0 検索を担当します。</li>
          <li><strong>Redis 通信バス</strong>：Pub/Sub メカニズムを利用して、ノード間のリアルタイム通信とタスク配信を実現します。</li>
        </ul>
        <h3>Global Agent Mesh の特徴</h3>
        <p>Mesh ネットワークのハブとして、BotNexus はドメインを越えた発見、フェデレーション認証、B2B 連携ロジックを処理し、異なる企業のエージェント間での安全かつ効率的な機能補完を保証します。</p>
      `
    },
    2: {
      title: 'インテリジェント Swarm エンジン',
      category: '高度な技術',
      date: '2026-01-11',
      summary: 'Swarm エンジンを利用してマルチエージェント連携と動的タスク分解を実現する方法を学びます。',
      content: `
        <h3>Swarm エンジンとは？</h3>
        <p>Swarm エンジンは、BotMatrix が複雑なタスクを処理するための連携フレームワークです。大きな目標を複数のサブタスクに分解し、最適な専門エージェントに割り当てて実行します。</p>
        <h4>主要なメカニズム</h4>
        <ul>
          <li><strong>タスク分解器</strong>：LLM を利用して複雑な指示を順序付けられたタスクフローに変換します。</li>
          <li><strong>動的ハンドオフ</strong>：エージェントは処理結果に基づいてタスクの制御権を動的に移譲できます。</li>
          <li><strong>集合知</strong>：マルチエージェントによる投票や合意形成メカニズムを通じて意思決定の精度を高めます。</li>
        </ul>
        <p>2026年第3四半期のロードマップでは、Swarm エンジンはさらに複雑なタスクオーケストレーションと分散型合意アルゴリズムをサポートする予定です。</p>
      `
    },
    3: {
      title: 'コンピュータ使用 (Computer Use) ガイド',
      category: '先端探索',
      date: '2026-01-12',
      summary: 'エージェントが人間のようにデスクトップ環境や Web アプリを操作する方法を探ります。',
      content: `
        <h3>エージェントに「手」を与える</h3>
        <p>Computer Use 技術により、BotMatrix のエージェントはクリック、ドラッグ、入力、視覚解析など、OS と直接対話できるようになります。</p>
        <h4>活用シーン</h4>
        <ul>
          <li><strong>自動テスト</strong>：クロスプラットフォームの UI 回帰テストを自動的に完了します。</li>
          <li><strong>データ収集</strong>：API のないレガシーシステムから構造化データを抽出します。</li>
          <li><strong>プロセス自動化</strong>：複数のソフトウェア間にまたがる人間のワークフローをシミュレートします。</li>
        </ul>
        <p>この機能は現在 BotWorker 拡張モジュールに統合されており、Linux および Windows のヘッドレスモードでの実行をサポートしています。</p>
      `
    },
    4: {
      title: 'デジタル従業員システム設定',
      category: '製品ソリューション',
      date: '2026-01-13',
      summary: '認知、アイデンティティ、KPI を備えた独立した生産性ユニットを構築する方法。',
      content: `
        <h3>デジタル従業員の概念</h3>
        <p>単なるチャットボットではなく、デジタル従業員は完全なアイデンティティ識別、権限体系、業務能力を備えたエージェントであり、システム内では一意の IdentityGORM で識別されます。</p>
        <h4>コア機能</h4>
        <ul>
          <li><strong>インテントスケジューラ</strong>：ユーザーのニーズと従業員のスキルを正確にマッチングさせます。</li>
          <li><strong>操作監査</strong>：エージェントの操作パスを完全にログ記録し、コンプライアンスと追跡可能性を確保します。</li>
          <li><strong>スキルセット (Toolset)</strong>：MCP プロトコルを通じて動的にロードされるビジネス機能パッケージ。</li>
        </ul>
      `
    },
    5: {
      title: '認知メモリと RAG 2.0',
      category: '人工知能',
      date: '2026-01-14',
      summary: 'ベクトルデータベースとナレッジグラフに基づく永続メモリソリューション。',
      content: `
        <h3>検索から認知へ</h3>
        <p>BotMatrix の RAG 2.0 は、単なるベクトル検索を超え、Agentic RAG と GraphRAG 技術を導入しています。</p>
        <h4>技術的なハイライト</h4>
        <ul>
          <li><strong>短期メモリ</strong>：Redis に基づくスライディングウィンドウ・コンテキスト管理。</li>
          <li><strong>長期メモリ</strong>：ベクトルデータベースに基づく意味検索。</li>
          <li><strong>ナレッジグラフ</strong>：Neo4j などのグラフデータベースを使用して、ドキュメントをまたがる複雑な関係推論を処理します。</li>
        </ul>
      `
    },
    6: {
      title: 'MCP プラグイン開発マニュアル',
      category: '開発者センター',
      date: '2026-01-15',
      summary: 'Model Context Protocol 標準に従い、エージェント機能を迅速に拡張します。',
      content: `
        <h3>MCP 標準プロトコル</h3>
        <p>Model Context Protocol (MCP) は、BotMatrix が推奨するプラグイン開発標準であり、機能提供者と消費者の疎結合を実現します。</p>
        <h4>開発ステップ</h4>
        <ul>
          <li><strong>リソースの定義</strong>：静的データや動的ドキュメントを公開します。</li>
          <li><strong>ツールの公開</strong>：LLM から呼び出し可能な関数インターフェースを提供します。</li>
          <li><strong>プロンプトのプリセット</strong>：よく使われるビジネスプロンプトテンプレートをカプセル化します。</li>
        </ul>
        <p>stdio、SSE、WebSocket などの転送方法をサポートし、Python、Go、Node.js などの主要な言語と互換性があります。</p>
      `
    }
  }
};
