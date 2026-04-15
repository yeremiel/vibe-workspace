# CLAUDE.md — 워크스페이스 가이드

> AI 어시스턴트(Claude Code)가 이 워크스페이스에서 작업할 때 참조하는 메타 문서.

---

## 1. 작업자 정보 (WHO)

| 항목 | 값 |
|------|-----|
| 이름 | 미엘 (yeremiel) |
| 주 사용 언어 | Java (16년차), Go, TypeScript, Flutter |
| 타임존 | Asia/Seoul (KST, UTC+9) |
| 대화 언어 | 한국어 |

---

## 2. 디렉터리 구조 (WHERE)

```
./
├── knowledge_base/                  # 지식 베이스 (예정)
├── obsidian/                        # Obsidian 볼트 (지식 베이스 · 작업 기록)
├── projects/
│   └── self/
│       ├── claw-dash/               # OpenClaw Workbench (repo 식별자: claw-dash)
│       │   └── CLAUDE.md
│       ├── claw-usage-chart/        # OpenClaw 토큰 사용량 시각화 도구 (→ claw-dash 통합 예정)
│       │   └── CLAUDE.md
│       ├── kanban-test/              # 칸반 프로젝트 (실험, 초기 init만)
│       ├── Learning-Go-study/       # Go 언어 학습 프로젝트 (ch1~ch2)
│       ├── scanner/                 # 오프라인 문서 스캐너 앱 (Flutter + Android CameraX/OpenCV)
│       └── myLib/                   # 개인 도서 관리 서비스 (메인 프로젝트)
│           ├── CLAUDE.md            # 워크스페이스 오케스트레이션
│           ├── docs/                # 워크스페이스 레벨 문서
│           ├── myLib-backend/       # Go 백엔드 API (독립 git repo)
│           │   └── CLAUDE.md        # Go/Gin/MongoDB/Redis 규칙
│           └── myLib-frontend/      # Flutter 프론트엔드 앱 (독립 git repo)
│               └── CLAUDE.md        # Flutter/Riverpod/Dio 규칙
├── shared/                          # 프로젝트 간 공유 리소스 (예정)
└── CLAUDE.md                        # ← 이 파일
```

---

## 3. 개발 환경 (WHERE)

### 시스템

| 항목 | 버전 |
|------|------|
| OS | macOS 26.3.1 (Darwin 25.3.0 arm64) |
| 아키텍처 | Apple Silicon (arm64) |

### 런타임 / SDK

| 도구 | 설치 버전 | 비고 |
|------|-----------|------|
| Go | 1.26.1 | myLib-backend: `go 1.25.5` (go.mod) |
| Flutter | 3.38.7 (stable) | myLib-frontend, scanner |
| Dart | 3.10.7 | Flutter 내장 |
| Node.js | 25.8.2 | |
| npm | 11.11.1 | |
| Bun | 1.3.11 | |
| pnpm | 10.33.0 | |
| Docker | — | 미설치 |
| mongosh | 2.8.2 | MongoDB 셸 |
| Java | OpenJDK 25.0.2 (Temurin) | scanner Android 빌드용 |

### 주요 프로젝트별 기술 스택

| 프로젝트 | 언어 | 프레임워크 / 핵심 라이브러리 | DB / 인프라 |
|----------|------|------------------------------|-------------|
| myLib-backend | Go 1.25 | Gin, JWT v5, Viper, Zap, Swaggo | MongoDB 6.0+, Redis 7.0+ |
| myLib-frontend | Dart 3.10 / Flutter 3.38 | Riverpod 3.x, Dio 5.x, go_router, Freezed | flutter_secure_storage |
| claw-dash (OpenClaw Workbench) | Go + Svelte 5 | Gin (BFF), Vite | OpenClaw Gateway (WebSocket) |
| claw-usage-chart | Go | embed.FS, modernc.org/sqlite | SQLite (WAL), Chart.js — 영어 커밋/주석 컨벤션, claw-dash 통합 예정 |
| scanner | Dart 3.10 / Flutter 3.38 | CameraX, OpenCV (Android NDK) | 로컬 저장 (MediaStore) |
| Learning-Go-study | Go 1.22 | 표준 라이브러리 | — |

---

## 4. 작업 규칙 (HOW)

### 언어 및 주석
- 모든 대화, 주석, 커밋 메시지는 **한국어**로 작성
- 코드 식별자(변수, 함수명 등)는 영어 유지

### Git 워크플로우
- 커밋 메시지 컨벤션, 브랜치 전략 → `knowledge_base/common/git-workflow.md` 참조
- 커밋 전 반드시 메시지를 사용자에게 제안 → 승인 후 커밋

### Knowledge Base (공통 개발 표준)

`knowledge_base/` 디렉터리에 프로젝트 횡단 개발 표준이 정의되어 있다.

**모든 프로젝트 공통 적용:**

| 문서 | 경로 | 내용 |
|------|------|------|
| API 응답 표준 | `knowledge_base/common/api-response.md` | 응답 envelope 구조, 에러 코드 규칙, HTTP 상태 코드 매핑 |
| Git 워크플로우 | `knowledge_base/common/git-workflow.md` | 브랜치 전략, 커밋 메시지 컨벤션, 커밋 전 승인 절차 |

**언어별 문서**(`go/`, `flutter/`, `java/` 등)는 각 프로젝트 CLAUDE.md에서 해당하는 것만 참조한다.

### 프로젝트별 상세 규칙
- 각 프로젝트의 `CLAUDE.md`에 코딩 규칙, 아키텍처 패턴, 테스트 가이드가 정의됨
- 이 파일에서 중복하지 않고, **하위 CLAUDE.md에 위임**
- 프로젝트별 Knowledge Base 예외 사항은 해당 프로젝트 CLAUDE.md에 명시


---

## 5. 금지사항 (WHAT NOT)

| # | 규칙 | 설명 |
|---|------|------|
| 1 | `.env` / API 키 하드코딩 금지 | 환경 변수 또는 설정 파일을 통해 관리 |
| 2 | 생성물 직접 수정 금지 | `build/`, `node_modules/`, `.g.dart`, `generated/` 등 빌드 산출물 수동 편집 금지 |
| 3 | 자의적 작업 실행 금지 | 지시하지 않은 리팩토링, 기능 추가, 파일 삭제 등 임의 실행 금지 |
| 4 | 캐시 탓 하지 말 것 | 문제 발생 시 캐시가 아닌 자기 코드를 먼저 의심할 것 |
| 5 | 불가능 즉시 선언 금지 | 불가능 판단 전에 **최소 3가지 대안**을 제시할 것 |
| 6 | 추측성 URL 생성 금지 | 확인되지 않은 URL을 생성하거나 제안하지 않을 것 |
| 7 | 민감 정보 로깅 금지 | 토큰, 비밀번호 등 민감 데이터를 로그에 남기지 않을 것 |

---

## 6. 프로젝트별 빠른 참조

### myLib-backend

```bash
make run               # 빌드 후 실행 (Swagger 포함)
make test              # 전체 테스트
make lint              # golangci-lint
make swagger           # Swagger 문서 생성
```

### myLib-frontend

```bash
make deps              # flutter pub get
make build-runner      # 코드 생성 (freezed, json_serializable)
make test              # 유닛 테스트
make api-update        # 백엔드 Swagger → Dart 클라이언트 재생성
```

### claw-dash

```bash
make dev               # 백엔드+프론트 동시 실행
make ci                # 빌드/테스트/보안/MVP 검증 통합
make security          # 정적 점검
```

### claw-usage-chart

```bash
go build -o claw-usage-chart .   # 빌드
./claw-usage-chart               # 실행 (localhost:8585)
./claw-usage-chart -d            # 데몬 모드
```

### scanner

```bash
flutter test                     # 유닛 테스트
flutter analyze                  # 정적 분석 (JAVA_HOME 설정 필요)
flutter build apk --debug        # 디버그 APK 빌드 (JAVA_HOME 설정 필요)
```

### 공통 아키텍처

```
Backend:  Handler → Service → Repository (Clean Architecture)
Frontend: Screen/Widget → Provider → Repository → DataSource (Feature-First)
API 연동: Backend swagger.json → Frontend make api-update → 자동 생성 클라이언트
```

---

*마지막 갱신: 2026-04-05*
