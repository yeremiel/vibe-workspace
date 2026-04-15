# Flutter/Dart 주의사항

## `copyWith()`에서 nullable 필드를 `null`로 리셋할 수 없음

Dart의 `copyWith()` 패턴에서 `field: field ?? this.field`는 `null`을 "변경 없음"으로 취급한다.
nullable 필드를 `null`로 초기화해야 하는 경우 `copyWith()` 대신 `const State()` 새 인스턴스를 생성한다.

```dart
// ❌ meta가 null로 리셋되지 않음
state = state.copyWith(meta: null, books: []);

// ✅ 새 인스턴스 생성으로 해결
state = const BookSearchState();
```

## 생성된 API 모델과 프로젝트 모델의 Import 충돌

Swagger 생성 코드에 `User` 등 프로젝트 모델과 동일한 이름이 존재할 수 있다.
충돌 시 generated export 파일에서 `hide`로 제외한다.

```dart
// ❌ User 클래스 충돌
import 'package:mylib_app/core/api/generated/export.dart';

// ✅ hide로 충돌 해결
import 'package:mylib_app/core/api/generated/export.dart' hide User;
```

## Riverpod Notifier에서 `state` 이름 충돌

Notifier 클래스 내부에서 `state`는 Riverpod의 상태 프로퍼티이다.
OAuth 등 외부 파라미터에 `state`라는 이름이 있으면 충돌하므로 별도 이름을 사용한다.

```dart
// ❌ Riverpod state 프로퍼티와 충돌
final state = generateOAuthState();

// ✅ 명시적 이름으로 구분
final oauthState = generateOAuthState();
```

## `.g.dart` 파일은 git에 포함하지 않음

`build_runner`가 생성하는 `.g.dart` 파일은 `.gitignore`에 등록되어 있다.
`git add` 시 자동으로 제외되지만, 수동 추가하지 않도록 주의한다.
