# knowledge_base — 生きた開発標準

🇯🇵 日本語 | [🇰🇷 한국어](README.md)

人間とAIエージェント（Claude Code）が同じドキュメントを読み、同じルールで働くための開発標準集です。

ワークスペース全体のthesis → [最上位README](../README.ja.md)

## どう機能するか

各プロジェクトの `CLAUDE.md` から `@` 参照で必要な標準ドキュメントだけを読み込みます。

```markdown
# プロジェクト CLAUDE.md の例

このプロジェクトは次の標準に従う。

@../../knowledge_base/common/api-response.md
@../../knowledge_base/common/git-workflow.md
@../../knowledge_base/go/conventions.md
@../../knowledge_base/go/design-patterns.md
```

この一度の宣言で:

- AIエージェントはコード生成時に常にその標準に従う
- 人間（自分や協働者）も同じドキュメントでオンボーディング
- プロジェクトが変わっても同じ送受信仕様・同じ規約が保たれる

## ディレクトリ構成

### `common/` — 言語非依存の標準

全プロジェクトに共通適用。バックエンドでもフロントエンドでも同じ。

| ドキュメント | 内容 |
|--------------|------|
| [`api-response.md`](common/api-response.md) | バックエンドAPIのレスポンスエンベロープ、エラーコード体系、HTTPステータスマッピング。**言語/フレームワーク非依存** |
| [`git-workflow.md`](common/git-workflow.md) | ブランチ戦略、コミットメッセージ規約、コミット前の承認手順 |

### `go/` — Goバックエンド標準

| ドキュメント | 内容 |
|--------------|------|
| [`conventions.md`](go/conventions.md) | パッケージ・関数・変数の命名、エラー処理、ロギング、テスト作成規約 |
| [`design-patterns.md`](go/design-patterns.md) | インターフェース活用、Clean Architectureの階層分離、DI、並行性パターン |

### `flutter/` — Flutter/Dart標準

| ドキュメント | 内容 |
|--------------|------|
| [`conventions.md`](flutter/conventions.md) | ファイル・ウィジェット・状態の命名、ルーティング、ディレクトリ構造 |
| [`design-patterns.md`](flutter/design-patterns.md) | Riverpod 3.x Notifierパターン、Feature-First構成、Freezedの活用 |
| [`gotchas.md`](flutter/gotchas.md) | 作業中に遭遇した落とし穴と解決方法の蓄積（`copyWith` のnullable処理など） |

## 「生きている」ということ

標準は**一度作って終わる静的なドキュメントではありません。**

- 新しいプロジェクトで見つけたパターンが標準に昇格する
- 繰り返し遭遇したミスが `gotchas.md` のようなドキュメントに蓄積される
- 新しい言語/ドメインが現れれば、このディレクトリ配下に新しいフォルダが生まれる（例: 将来的な `java/`、`typescript/`）

この進化のプロセス自体がワークスペースの成果物です。claw-dashケーススタディでは、BFFの特性により `api-response.md` の一部原則を意図的に例外として扱った事例を扱っています → [cases/claw-dash/](../cases/claw-dash/)

## どこから読むのが良いか

初めて訪れた方には、次の順序をおすすめします。

1. **[`common/api-response.md`](common/api-response.md)** — 最も強制力のある標準。「自分が作るすべてのバックエンドAPIはこう応答する」という約束の実体
2. **[`common/git-workflow.md`](common/git-workflow.md)** — 短く直感的。作業フローの前提
3. 関心のあるスタックの `conventions.md` → `design-patterns.md` の順
4. （Flutterなら）[`flutter/gotchas.md`](flutter/gotchas.md) — 生きたシステムの最も直接的な証拠
