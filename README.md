# spider

Knuth & Ruskey의 논문 **["Efficient Coroutine Generation of Constrained Gray Sequences"](https://www-cs-faculty.stanford.edu/~knuth/papers/spider.pdf)** (Ole-Johan Dahl 추모 헌정)에 나오는 알고리즘들을 Go로 충실히 구현한 프로젝트입니다.

저자들이 익살스럽게 **"spider squishing"** 이라 부르는 문제를 다룹니다.

## 무슨 문제인가

주어진 **방향 그래프**의 제약을 만족하는 모든 비트 문자열 $a_1 a_2 \dots a_n \in \{0,1\}^n$ 을 생성합니다. 제약은 간선 $j \to k$ 마다 $a_j \le a_k$ (즉 $a_j=1 \Rightarrow a_k=1$). 이는 형식적으로 **"비순환 poset의 order ideal을 모두 생성하는 문제"** 입니다.

여기에 더해 **그레이 경로(Gray path)** 로 — 한 단계에 정확히 비트 하나만 바뀌도록 — 나열합니다.

무향 그래프로 봐도 사이클이 없는 경우(_totally acyclic_)를 **거미(spider)** 라 부르며, 이때 그레이 경로가 항상 존재하고 **비트 변화당 상수 시간**에 생성할 수 있다는 것이 논문의 핵심 결과입니다.

```
  3   5          제약: a1≤a2≤a3, a4≤a3, a2≤a5 ...
   \ / \         정점을 0/1로 켜고 끄되,
    2   6   8    매 단계 한 비트만 바꾸며
     \  |  /     모든 유효 조합을 훑는다.
        1
```

## 두 개의 세계, 서로를 검증하다

논문은 같은 알고리즘을 두 가지 방식으로 제시하는데, 이 저장소는 **둘 다** 구현하고 **서로의 정답지로** 교차검증합니다.

| 세계 | 패키지 | 정체 |
|---|---|---|
| **협력하는 코루틴 (트롤)** | `poke` `bump` `nudge` | 각 트롤 $T_k$ 를 goroutine 하나로. 논문 의사코드를 거의 한 줄씩 옮긴 충실한 버전. |
| **active list** | `active` | 위 코루틴 무리를 명시적 자료구조 + 반복 루프로 "컴파일"한 효율 버전 (amortized $O(1)$). |

핵심 트릭은 `ret`/`invoke` 헬퍼로 **goroutine의 실행 위치(PC)가 곧 코루틴의 상태**가 되게 한 것입니다. 덕분에 명시적 상태 기계 없이 논문의 코루틴을 그대로 옮길 수 있습니다 (`nudge`는 사이클 중간 진입을 `goto` 한 줄로 처리).

그리고 두 세계의 출력이 특수 케이스에서 **패턴 단위로 완전히 일치**함을 테스트로 못 박았습니다 — 논문 §8의 주장("the three steps faithfully implement those coroutines")을 실증한 셈입니다.

## 패키지 구성

```
spider/
├── poke/      §1 무제약        — 표준 반사 그레이 코드 (2^n 패턴)
├── bump/      §2 체인          — 0 ≤ a1 ≤ … ≤ an ≤ 1 (n+1 패턴)
├── nudge/     §3 울타리        — a1 ≤ a2 ≥ a3 ≤ … (초기화가 까다로움)
├── spider/    B1 거미 데이터 모델 (자식/부호/scope, 근접집합 U_k/V_k, ideal 개수 n_k)
│              B2 §6 launching (초기 설정 α, 전이 τ, 최종 ω)
│              + brute-force 정답 열거기 (AllIdeals)
├── active/    §8 active list 생성기 — 임의의 거미를 amortized O(1)로
└── main.go    데모 드라이버
```

## 빠른 시작

```bash
go test ./...           # 전체 검증
go test -race ./...     # 코루틴 동시성 검증
```

### 코루틴 데모 (§1–3)

```bash
go run . -coro poke  -n 3        # 표준 그레이 코드
go run . -coro bump  -n 3        # 체인
go run . -coro nudge -n 4        # 울타리
```

```
$ go run . -coro bump -n 3
bump trolls — n=3  (chain: 0 <= a_1 <= … <= a_n <= 1)

000   initial state
001   bump{1,2,3} = true
011   bump{1,2,3} = true
111   bump{1,2} = true
111   bump{1} = false
---
...
```

출력의 패턴 열과 트롤 호출 집합이 논문 5·7·11쪽 표와 정확히 일치합니다.

### active list 데모 (§8)

```bash
go run . -coro active -spider example      # 논문 §4의 9정점 예제 거미 (60개 ideal)
go run . -coro active -spider chain  -n 5
go run . -coro active -spider fence  -n 6
```

```
$ go run . -coro active -spider example
active list — spider=example  (60 ideals, generated in Gray order)

000001100   1235679
000001101   1235679     ← 자는 노드는 터미널에서 밑줄로 표시
000001001   1235679
...
011011100   124679      ← P1 → Q1 전이 (48번째 패턴)
111011100   14789          양의 자식 2,6 이 음의 자식 8 로 교체
...
111111100   14789       ← 최종 상태
```

이 추적은 논문 21쪽 예제를 글자 단위로 재현합니다.

## 검증 내용

- **`poke`/`bump`/`nudge`** — 매 단계 1비트 변화, 모든 유효 패턴을 정확히 한 번씩, 그리고 논문 예제 표와 비트 단위 일치 (`-race` 클린).
- **`spider`** — 예제 거미의 `U_1={2,6,9}`, `V_1={4,7,8}`, scope, 개수 `n_k`(총 **60**), 그리고 §6 launch 표(α/τ/ω)가 논문 12·16·18쪽과 정확히 일치. brute-force 열거기로 개수 교차검증.
- **`active`** — NoArcs/Chain/Fence에서 `poke`/`bump`/`nudge`와 **출력 완전 일치**, 임의 혼합 거미에서 `AllIdeals`와 일치하는 완전한 그레이 코드, 논문 21쪽 추적 재현.

## 더 해볼 수 있는 것

- **§9 loopless $O(1)$** — focus pointer + lazy family update (현재 `active`의 삽입은 정렬 스캔).
- **일반 `gen` 코루틴 (§5)** — `maxu`/`maxv`/`prev` 테이블로 poke/bump/nudge를 통합하는 goroutine 버전.
- **TUI 시각화** — 거미와 active list를 실시간 애니메이션.

## 참고 문헌

- D. E. Knuth and F. Ruskey, *Efficient Coroutine Generation of Constrained Gray Sequences*. In _From Object-Orientation to Formal Methods: Essays in Memory of Ole-Johan Dahl_, LNCS 2635 (2004), 183–204.
- Y. Koda and F. Ruskey, *A Gray code for the ideals of a forest poset*, Journal of Algorithms **15** (1993), 324–340.
- D. E. Knuth, *The Art of Computer Programming*, Vol. 4, §7.2.1.1 (Generating all $n$-tuples).
