# Flutter 디자인 패턴

## Riverpod 패턴 (3.x)

- **Notifier 패턴 사용** (StateNotifier 사용 금지)
- 동기 상태: `Notifier` + `NotifierProvider`
- 비동기 상태: `AsyncNotifier` + `AsyncNotifierProvider`
- 상태 클래스는 immutable + `copyWith()` 메서드 제공

## Repository 패턴

- 인터페이스: `domain/repositories/`에 정의
- 구현체: `data/repositories/`에 구현
- DataSource와 Storage를 조합하여 데이터 접근 추상화

## 의존성 주입

- Riverpod Provider를 통해 의존성 주입
- DataSource → Repository → Notifier 순서로 Provider 정의

## 레이어 역할

- **Presentation**: UI 렌더링, 사용자 입력 처리
- **Provider**: 상태 관리, UI와 비즈니스 로직 연결
- **Repository**: 데이터 소스 추상화, 캐싱 전략
- **DataSource**: 실제 API 호출, 로컬 DB 접근
- **Model**: 데이터 구조 정의, 직렬화/역직렬화
