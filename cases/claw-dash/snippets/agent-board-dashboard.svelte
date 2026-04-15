<!-- 발췌 출처: claw-dash frontend/src/components/board/AgentBoardDashboard.svelte -->
<!-- 전체 코드는 비공개입니다. 핵심 로직 발췌. -->
<!--
  Svelte 5 Runes 기반 에이전트 칸반 보드 컴포넌트.
  - $props(), $state(), $derived() 등 Runes API 사용
  - boardState (class-based reactive store) 구독 및 렌더링
  - parent/child 카드 관계 강조, guardrail 배지, 우선순위 표시
  - CSS 토큰 기반 테마 (color-mix() 활용)
-->

<script lang="ts">
  import { onMount } from 'svelte';
  import { projectsState } from '../../lib/state/projects.svelte';
  import { teamsViewState } from '../../lib/state/teams.svelte';
  import {
    boardState,
    BOARD_STATUSES,
    BOARD_STATUS_LABELS,
    type BoardCard,
    type BoardStatus
  } from '../../lib/state/board.svelte';

  interface GuardrailBadge {
    label: string;
    tone: 'stale' | 'review' | 'decision';
  }

  interface Props {
    onOpenCreateModal?: () => void;
    onOpenDetailModal?: () => void;
  }

  // Svelte 5 Runes: props 선언
  let {
    onOpenCreateModal = () => {},
    onOpenDetailModal = () => {}
  }: Props = $props();

  const PRIORITY_LABELS: Record<string, string> = {
    p0: 'P0', p1: 'P1', p2: 'P2', p3: 'P3'
  };

  // Svelte 5 Runes: 반응형 상태
  let highlightedParentId = $state<string | null>(null);

  // Svelte 5 Runes: 파생 상태 (자동으로 의존성 추적)
  const selectedCard = $derived(boardState.detailCard);

  onMount(() => {
    void boardState.initialize();
    void projectsState.refreshProjectContext();
    if (!teamsViewState.loaded && !teamsViewState.loading) {
      void teamsViewState.refresh();
    }

    // 20초마다 카드 목록 및 선택된 카드 상세를 자동 갱신
    const timer = window.setInterval(() => {
      void boardState.refreshCards();
      if (boardState.selectedCardId && !boardState.actionLoading) {
        void boardState.refreshDetail();
      }
    }, 20000);

    return () => window.clearInterval(timer);
  });

  // --- 카드 관계 강조 로직 ---

  function isHighlightedParent(card: BoardCard): boolean {
    return highlightedParentId === card.id;
  }

  function isHighlightedChild(card: BoardCard): boolean {
    return highlightedParentId !== null && card.parentCardId === highlightedParentId;
  }

  function isDimmedByRelation(card: BoardCard): boolean {
    if (!highlightedParentId) return false;
    return !isHighlightedParent(card) && !isHighlightedChild(card);
  }

  function toggleChildHighlight(cardId: string): void {
    highlightedParentId = highlightedParentId === cardId ? null : cardId;
  }

  // --- guardrail 배지 ---

  function guardrailsFor(card: BoardCard): GuardrailBadge[] {
    const badges: GuardrailBadge[] = [];
    if (card.guardrails.stale)         badges.push({ label: 'Stale',           tone: 'stale'    });
    if (card.guardrails.reviewNeeded)  badges.push({ label: 'Review needed',   tone: 'review'   });
    if (card.guardrails.needsDecision) badges.push({ label: 'Needs decision',  tone: 'decision' });
    return badges;
  }

  function formatPriority(value: string): string {
    const normalized = value.trim().toLowerCase();
    if (!normalized) return '—';
    return PRIORITY_LABELS[normalized] ?? normalized.toUpperCase();
  }

  function selectCard(cardId: string): void {
    highlightedParentId = null;
    void boardState.selectCard(cardId);
    onOpenDetailModal();
  }
</script>

<section class="board-dashboard">
  {#if boardState.error}
    <p class="feedback error">{boardState.error}</p>
  {/if}

  <section class="workspace">
    <section class="board-stage">
      <div class="desktop-board" aria-label="Agent board columns">
        <!--
          BOARD_STATUSES: ['inbox', 'planned', 'in_progress', 'review', 'done', 'blocked']
          각 컬럼을 순회하며 해당 status의 카드를 렌더링
        -->
        {#each BOARD_STATUSES as status}
          <section class={`column ${status}`}>
            <header class="column-header">
              <div class="column-header-main">
                <span class={`summary-dot ${status}`}></span>
                <h3>{BOARD_STATUS_LABELS[status]}</h3>
              </div>
              <small>{boardState.count(status)}</small>
            </header>

            <div class="column-body">
              {#if status === 'inbox'}
                <button type="button" class="intake-create-card" onclick={onOpenCreateModal}>
                  <span class="plus">＋</span>
                  <span>Add to Inbox</span>
                </button>
              {/if}

              {#each boardState.cardsForStatus(status) as card (card.id)}
                <button
                  type="button"
                  class="card-tile"
                  class:selected={selectedCard?.id === card.id}
                  class:relation-parent={isHighlightedParent(card)}
                  class:relation-child={isHighlightedChild(card)}
                  class:dimmed={isDimmedByRelation(card)}
                  onclick={() => selectCard(card.id)}
                >
                  <div class="tile-top">
                    <span class={`status-chip ${status}`}>{BOARD_STATUS_LABELS[card.status]}</span>
                    {#if card.priority.trim()}
                      <span class="meta-chip">{formatPriority(card.priority)}</span>
                    {/if}
                  </div>

                  <div class="tile-copy">
                    <strong>{card.title}</strong>
                  </div>

                  <div class="tile-meta">
                    <span>{card.teamLabel || 'Unassigned'}</span>
                  </div>

                  <!-- parent/child 관계 토글 -->
                  {#if card.parentCardId.trim().length > 0}
                    <div class="relation-chip child-relation">Child</div>
                  {:else if boardState.cards.filter(c => c.parentCardId === card.id).length > 0}
                    <span
                      class="relation-chip parent-relation relation-toggle"
                      role="button"
                      tabindex="0"
                      onclick={(event) => { event.stopPropagation(); toggleChildHighlight(card.id); }}
                      onkeydown={(event) => {
                        if (event.key === 'Enter' || event.key === ' ') {
                          event.preventDefault();
                          event.stopPropagation();
                          toggleChildHighlight(card.id);
                        }
                      }}
                    >
                      Parent
                    </span>
                  {/if}

                  <!-- guardrail 배지: 오래된 카드, 리뷰 필요, 의사결정 대기 -->
                  {#if guardrailsFor(card).length > 0}
                    <div class="guardrails-row">
                      {#each guardrailsFor(card) as badge}
                        <span class={`guardrail-badge ${badge.tone}`}>{badge.label}</span>
                      {/each}
                    </div>
                  {/if}

                  {#if card.nextAction}
                    <div class="tile-note">
                      <span>Next</span>
                      <p>{card.nextAction}</p>
                    </div>
                  {/if}
                </button>
              {/each}
            </div>
          </section>
        {/each}
      </div>
    </section>
  </section>
</section>

<style>
  /* CSS 토큰 기반 테마 — 하드코딩 색상 없음 */
  .board-dashboard {
    position: relative;
    isolation: isolate;
    padding: 0.7rem;
    display: flex;
    flex-direction: column;
    gap: 0.9rem;
    min-height: 100%;
    height: 100%;
  }

  /* 배경: 다크/라이트 테마 자동 대응 (CSS 토큰 + color-mix) */
  .board-dashboard::before {
    content: '';
    position: absolute;
    inset: 0;
    pointer-events: none;
    background:
      radial-gradient(circle at 86% 10%, color-mix(in srgb, var(--accent-blue) 12%, transparent), transparent 28%),
      radial-gradient(circle at 14% 88%, color-mix(in srgb, var(--accent-green) 10%, transparent), transparent 32%),
      radial-gradient(circle at 48% 46%, color-mix(in srgb, var(--selected-bg) 16%, transparent), transparent 36%);
    filter: blur(26px);
    opacity: 0.84;
  }

  /* status별 색상 */
  .summary-dot.inbox,    .status-chip.inbox    { background: color-mix(in srgb, #6fc2ff 72%, var(--fill-base)); }
  .summary-dot.planned,  .status-chip.planned  { background: color-mix(in srgb, #ffd36d 76%, var(--fill-base)); }
  .summary-dot.in_progress, .status-chip.in_progress { background: color-mix(in srgb, #7ee7b8 76%, var(--fill-base)); }
  .summary-dot.review,   .status-chip.review   { background: color-mix(in srgb, #c9a7ff 72%, var(--fill-base)); }
  .summary-dot.done,     .status-chip.done     { background: color-mix(in srgb, var(--accent-green) 72%, var(--fill-base)); }
  .summary-dot.blocked,  .status-chip.blocked  { background: color-mix(in srgb, var(--danger) 72%, var(--fill-base)); }

  /* guardrail 배지 */
  .guardrail-badge.stale    { background: color-mix(in srgb, var(--accent-orange) 18%, var(--fill-base)); color: var(--accent-orange); }
  .guardrail-badge.review   { background: color-mix(in srgb, #c9a7ff 18%, var(--fill-base));              color: #c9a7ff; }
  .guardrail-badge.decision { background: color-mix(in srgb, var(--danger) 14%, var(--fill-base));        color: var(--danger); }

  .card-tile {
    border: 1px solid var(--panel-border);
    background: var(--panel);
    color: var(--text-primary);
    border-radius: 10px;
    padding: 0.7rem 0.75rem;
    width: 100%;
    text-align: left;
    cursor: pointer;
    transition: border-color 180ms;
  }

  .card-tile:hover                { border-color: var(--border-interactive); }
  .card-tile.selected             { border-color: var(--selected-border); background: var(--selected-bg); }
  .card-tile.relation-parent      { border-color: var(--accent-blue); }
  .card-tile.relation-child       { border-color: color-mix(in srgb, var(--accent-blue) 50%, transparent); }
  .card-tile.dimmed               { opacity: 0.38; }
</style>
