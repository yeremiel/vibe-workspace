# OpenClaw Workbench — ケーススタディ

> AIエージェントランタイム（OpenClaw Gateway）を運用するためのダッシュボード兼オーケストレーションツール。
> Go BFF + Svelte 5 SPA構成で、MVPからKanbanベースのエージェント作業ボードまで7つのPhaseを完走した。

🇯🇵 日本語 | [🇰🇷 한국어](README.md)

---

## 1. 背景 — 「なぜ作ったのか」

OpenClaw GatewayはAIエージェントを実行するランタイムだが、セッションの状態を一目で確認したり、エージェントに指示を出すには、毎回CLIとJSON-RPCを直接操作する必要があった。セッションが増えるにつれ、複数のターミナルを開いて手動で状態を確認する方法には明確な限界があった。

必要なものは3つだった。第一に、エージェントセッションをリアルタイムに監視できるダッシュボード。第二に、チャットでエージェントに指示を送り応答を受け取るインターフェース。第三に、複数のエージェント作業をKanbanボードでオーケストレーションするツール。この3つを1つのWebアプリにまとめたのがOpenClaw Workbenchだ。

---

## 2. AIコンテキスト設計 — 「AIにどう作業を任せたか」

このプロジェクトではClaude Codeをメインの実装ツールとして使用した。AIがプロジェクトの文脈を正確に把握できるよう、CLAUDE.mdを階層化したことが核心だ。

- **ワークスペース CLAUDE.md**: 作業者情報、ディレクトリ構造、共通禁止事項、言語ルールなど、全プロジェクトに適用される基盤コンテキスト
- **knowledge_base/go/**: GoのコーディングコンベンションとClean Architectureパターンのドキュメント。`@`参照で必要な時だけロードしてコンテキストの無駄を削減
- **claw-dash/CLAUDE.md**: プロジェクト固有のルール。BFFの特性上、Response Envelopeを適用しないこと、Gatewayの障害時にHTTP 200 fallbackを返すといった標準の例外事項を明記（[CLAUDE.md 参照](./CLAUDE.md)）

作業サイクルは**設計ドキュメント作成 → 実装プラン生成 → ステップ実行**を繰り返した。

役割分担は以下の通り。

| 役割 | 担当 |
|------|------|
| 機能要件の定義・設計方針の決定 | 自分 |
| トレードオフの検討・設計の承認 | 自分 |
| 設計ドキュメントの初稿作成 | AI（Claude Code） |
| 実装プランの生成・ステップ実行 | AI（Claude Code） |
| コードレビュー・動作検証・回帰確認 | 自分 |
| 例外判断・設計変更の意思決定 | 自分 |バックエンド構造リファクタリングのように影響範囲が広い作業は、必ずデザインドキュメントを先に作成・承認してから実行した（[backend-structure-refactor-design](./plans/2026-03-20-backend-structure-refactor-design.md)）。Kanban安定化のような複雑なドメインでは、claim matrixで現状を事実と仮説に分類し、Phaseごとにスコープをロックして意味の変更と構造変更が混在しないよう制御した（[kanban-stabilization-plan](./plans/2026-03-24-kanban-stabilization-plan.md)）。プロジェクト全体で50本以上の計画ドキュメントが生成された。このうち代表的なものを[`plans/`](./plans/)に公開している。

---

## 3. 開発の流れ — 「実際にどう進めたか」

### 技術選定

**BFFにNodeではなくGoを選んだ。** myLibのバックエンドですでにGo/Ginを使っており、BFFの核心的な役割であるWebSocket接続管理とSSE fan-outでは、goroutineがNodeのevent loopより直感的だった。最初にNodeで始めて後でGoに書き直すと同じAPIを2回実装することになるため、最初からGoを選択した。

**フロントエンドにはReactではなくSvelte 5を選んだ。** AngularとJSPの経験者にとって、SvelteのHTML-firstテンプレートはReact JSXより自然だった。`$state`、`$derived`、`$effect` Runesで状態を管理することで、useState/useEffectフックのクロージャの落とし穴がない。何より、未経験の技術を試すこと自体が個人プロジェクトの目的だった（[agent-board-dashboard.svelte 参照](./snippets/agent-board-dashboard.svelte)）。

### アーキテクチャの進化

初期MVPでは`cmd/server/main.go`に依存性の組み立てとルーティングが集中していた。機能が急速に追加されるにつれ、ハンドラーも単一ディレクトリに集積し、ビジネスロジックが`internal/kanban`、`internal/usage`、`internal/gateway`に散在して責務の境界が曖昧になった。

これを`Handler → Service → Repository`の階層にリファクタリングした。main.goをシンプルに保ち、アプリ組み立ての責務を`internal/app`に移動し、ドメインごとに`router`、`handler`、`service`を分離した。既存のAPIパスとJSONレスポンスのshapeは変更しないことを原則とした（[backend-structure-refactor-design 参照](./plans/2026-03-20-backend-structure-refactor-design.md)）。

### サーバー権威キューの導入

チャット入力の処理で重要な設計上の決定があった。フロントの状態だけで送信を制御すると、応答前に連続入力した場合にturnの競合が発生し、Stop直後の再送でrace conditionが起き、ブラウザリロードや複数クライアント環境で状態が不整合になる。

これを解決するために、セッションごとの単一in-flight + FIFOキューをサーバーで管理することにした。`/api/sessions/send`は`accepted`または`queued + queueIndex`を明示的に返し、フロントエンドが現在のメッセージの処理状態を正確に把握できるようにした（[session-chat-inline-message-design 参照](./plans/2026-03-19-session-chat-inline-message-design.md)）。

---

## 4. 試行錯誤 — 「詰まった時にどうしたか」

### バックエンド再起動失敗時に旧プロセスが動き続ける問題

コードを修正してバックエンドを再起動したのに動作が以前と同じだった。新プロセスがポートバインドに失敗して即座に終了していたが、旧バージョンのサーバーが生きているためヘルスチェックが正常に見えた、というのが原因だった。KanbanのPhase 1/2 Playwright検証で`finalReportPath`が空に見えた症状もこの問題だった（[kanban-phase1-2-playwright-validation 参照](./plans/2026-03-16-kanban-phase1-2-playwright-validation.md)）。

解決策は単純だった。ポート占有PIDの確認 → 終了 → 再起動 → 起動ログ/ヘルスチェック確認という順序をチェックリスト化した。教訓は「機能バグを疑う前にランタイム環境から確認する」ということだ。

### GatewayエラーがHTTP 200に隠れていた問題

`curl /api/sessions`のレスポンスが200 OKなのに、bodyに`{"error":"gateway rpc error..."}`が入っていた。Ginのログにも200しか記録されないため、エラーかどうかの判断がつかなかった。原因は、Gatewayの障害時もフロントエンドのfallbackのために常に200を返す設計だったためだ。

502への切り替えを試みたが、フロントエンドのパース構造と合わなかった。結局200を維持しつつ`slog.Warn`でGatewayの状態とエラーメッセージを出力する妥協案に落ち着いた（[gateway-client.go 参照](./snippets/gateway-client.go.txt)）。BFFパターンでは、アップストリームのエラーを透過的に伝えるか、fallbackで包むかを初期に決めておくべきだという教訓を得た。これを[CLAUDE.md](./CLAUDE.md)に「200 OK fallback」の例外事項として明記した。

### Kanbanオーケストレーションハブの結合度

Kanban機能がPhase 7まで進む中で、`completeRunSuccessLocked`という1つの関数がrun終了・artifact sync・review判定・parent refresh・auto-finishをすべて処理するようになった。単一関数のblast radiusが大きくなりすぎ、1箇所を修正すると連鎖的に影響が広がるようになった。

これを解決するために「behavior-preserving structural stabilization」戦略を採用した。新しいポリシーを追加せず、現在の意味を維持したままハブ関数をステップごとに分離するアプローチだ。claim matrixで事実と仮説を区別し、canonical semanticsの保存ルールを定義し、Phase 1〜4に分けて進めた（[kanban-stabilization-plan 参照](./plans/2026-03-24-kanban-stabilization-plan.md)）。

---

## 5. 結果 — 「何が出来上がったか」

### アーキテクチャ

```
┌───────────────────┐  REST / SSE   ┌──────────────┐  WebSocket   ┌──────────────────┐
│  Svelte 5 SPA     │ ───────────>  │  Go BFF      │  JSON-RPC    │  OpenClaw        │
│  (Vite + TS)      │ <───────────  │  (Gin :3001) │ <──────────> │  Gateway         │
│  :5173            │               └──────────────┘              └──────────────────┘
└───────────────────┘
     ブラウザ                           バックエンド                  エージェントランタイム
```

| レイヤー | 技術 |
|---------|------|
| フロントエンド | Svelte 5 (Runes) + Vite + TypeScript |
| バックエンド | Go + Gin |
| リアルタイム通信 | WebSocket (BFF ↔ Gateway)、SSE (ブラウザ ↔ BFF) |

### 主な機能

- **AIエージェントセッションのリアルタイム監視**: SSEベースのイベントストリーミング、keepaliveでプロキシ/ブラウザのタイムアウトを防止（[sse-handler.go 参照](./snippets/sse-handler.go.txt)）
- **Agent Board**: Kanbanベースの作業オーケストレーション。parent/childカードの関係強調、guardrailバッジ、優先度表示（[agent-board-dashboard.svelte 参照](./snippets/agent-board-dashboard.svelte)）
- **チャットインターフェース**: サーバー権威キューベースのメッセージ送信、インラインメッセージレンダリング
- **使用量分析ダッシュボード**: トークン使用量の可視化（SQLite + Chart.js）
- **Cronジョブ管理**: GatewayのCronタスクの作成・参照・削除

### 開発規模

- MVP（Phase 1〜5）+ Kanban/Agent Board（Phase 1〜7）完了
- 計画ドキュメント50本以上生成（代表的なものを`plans/`に公開）
- CSSトークンベースのダーク/ライトテーマシステム構築
- Gateway WebSocketクライアント: challengeハンドシェイク、exponential backoff再接続、並行安全RPC（[gateway-client.go 参照](./snippets/gateway-client.go.txt)）
