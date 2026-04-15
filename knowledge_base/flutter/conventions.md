# Flutter/Dart 코딩 컨벤션

## 명명 규칙

**파일 이름 (File Names)**:
- 모든 파일 이름은 `snake_case` 사용
- 클래스와 파일 이름 일치 (예: `LoginScreen` → `login_screen.dart`)
- **예시**: `auth_provider.dart`, `login_screen.dart`, `user_model.dart`

**클래스 이름 (Class Names)**:
- `PascalCase` 사용
- 역할을 명확히 표현하는 접미사 사용:
  - Screen: `LoginScreen`, `HomeScreen`
  - Widget: `BookCard`, `LoadingIndicator`
  - Provider/Notifier: `AuthNotifier`, `LibraryNotifier`
  - Repository: `AuthRepository`, `BookRepository`
  - Model: `User`, `Book`, `LoginRequest`
  - DataSource: `AuthRemoteDataSource`
- **예시**: `AuthNotifier`, `LibraryScreen`, `BookRepository`

**변수 및 함수 이름 (Variables & Functions)**:
- `lowerCamelCase` 사용
- 함수는 동사로 시작
- Boolean 변수/함수는 `is`, `has`, `can` 접두사 사용
- **예시**: `userName`, `isLoading`, `fetchBooks()`, `hasToken()`

**상수 (Constants)**:
- 최상위 상수: `lowerCamelCase` (Dart 공식 권장)
- 클래스 내 상수: `lowerCamelCase`
- **예시**: `const defaultTimeout = 30;`, `static const baseUrl = '...';`

**Provider 이름**:
- 접미사 `Provider` 사용
- 기능 + 역할 + Provider 형식
- **예시**: `authStateProvider`, `libraryBooksProvider`, `dioProvider`

## 에러 처리

- 모든 API 호출은 try-catch로 감싸기
- 커스텀 예외 클래스 사용 (`ApiException`)
- 에러 상태를 State에 포함하여 UI에서 처리
- 사용자에게 친화적인 에러 메시지 표시

## 위젯 작성

- 위젯은 단일 책임 원칙 준수
- 큰 위젯은 작은 위젯으로 분리
- `const` 생성자 적극 활용
- 비즈니스 로직은 위젯에서 분리 (Provider로 이동)

## Import 정리

`import` 문은 다음 순서로 정렬:
1. Dart SDK (`dart:`)
2. Flutter SDK (`package:flutter/`)
3. 외부 패키지 (`package:`)
4. 프로젝트 내부 패키지 (`package:<app>/`)
5. 상대 경로 (`../`, `./`)

## 로깅

- `debugPrint()` 사용 (릴리즈 빌드에서 자동 제거)
- 민감한 정보 로깅 금지 (토큰, 비밀번호 등)

## 성능

- `const` 위젯 적극 활용
- 불필요한 리빌드 방지 (`ref.watch` vs `ref.read` 적절히 사용)
- 이미지 캐싱 적용

## 테스트

**테스트 규칙**:
- Mock 라이브러리: `mocktail` 사용
- 테스트 파일명: `<대상>_test.dart` (예: `auth_provider_test.dart`)
- 그룹화: `group()`으로 관련 테스트 묶기
- 패턴: Arrange → Act → Assert

**테스트 대상**:
- Provider/Notifier: 상태 변경 및 메서드 동작
- Repository: 데이터 변환 및 에러 처리
- 핵심 위젯: 사용자 인터랙션 및 상태 표시

## 코드 품질 체크리스트

생성된 코드가 다음을 만족하는지 확인:
- [ ] 디렉토리 구조 준수 (Feature-First)
- [ ] 명명 규칙 일치 (snake_case 파일, PascalCase 클래스)
- [ ] Riverpod 3.x 패턴 사용 (Notifier, AsyncNotifier)
- [ ] 에러 처리 완료
- [ ] 테스트 코드 존재
- [ ] Import 그룹화 및 정렬
- [ ] const 생성자 활용
- [ ] 단일 책임 원칙 준수
- [ ] 레이어 분리 준수 (Presentation → Domain → Data)
