# Milestone 1 Context Tests (Simulation Core Stabilization)

## 目的

このドキュメントは、Simulation Core の前提・仕様・不変条件をテストとして固定し、将来の変更で壊れないようにするためのチェックリストです。  
テストは「仕様書」でもあるため、ここに書かれた内容と実装は常に一致している必要があります。

## 前提（Given）

- 路線は直線
- 固定閉塞（block occupancy）
- 列車は `block + progress(0.0..1.0) + direction` を持つ
- 速度は定数
- 駅はブロックの端（境界）として判定する
- シミュレーションは `Tick(dt)` により進行する（暗黙進行なし）

## テストカテゴリ

### A. 決定論（Determinism）

#### A-01: 同一入力で同一出力になる

- Given: 同じ初期 `SimulationState`（同じ列車・同じ占有・同じ時間）
- When: `Tick(dt)` を実行（同じ `dt`）
- Then: 返る Snapshot/State が完全一致する（列車位置、占有、時刻、イベント）

観点:
- `map iteration` による順序揺れがない
- 出力の列車並び順も安定している（DTO が `slice` なら順序固定）

### B. 駅境界モデル（Station as Boundary）

#### B-01: 列車は「駅にいる状態」を持たない

- Given: 列車は必ず `block` を持つ
- When: `Tick` や Snapshot を取得
- Then: 「駅専用の位置状態」が存在しない（position は常に block 上）

#### B-02: 駅判定は境界到達時のみ発生する

- Given: `0 < progress < 1`
- When: `Tick` 実行
- Then: 駅到着判定は発生しない（`atStation = false` / `boundaryStation = nil`）

#### B-03: 上り（Up）は `progress == 1.0` が駅境界

- Given: Up 方向、`progress` が `1.0` に到達
- When: `Tick`
- Then: 到着駅 = `ToStation(currentBlock)` として判定される

#### B-04: 下り（Down）は `progress == 0.0` が駅境界

- Given: Down 方向、`progress` が `0.0` に到達
- When: `Tick`
- Then: 到着駅 = `FromStation(currentBlock)` として判定される

#### B-05: 駅境界情報は Domain API で取得できる

- Given: 列車が境界にいる（progress 端）
- When: Domain API（例: `TrainStationAtBoundary(trainID)`）を呼ぶ
- Then: `(stationID, true)` が返る
- And: 境界でない場合は `(zero, false)` が返る

### C. Tick の基本挙動（進捗・遷移）

#### C-01: 境界未到達ならブロック内で進捗が増える

- Given: `0 < progress < 1`、次ブロック条件は問わない
- When: `Tick(dt)`
- Then: `block` は変わらない
- And: `progress` が増える（Up）/ 減る（Down）
- And: `0 < progress < 1` を保つ

#### C-02: 境界ちょうど到達（progress が端ぴったり）

- Given: Tick 後に `progress` がちょうど `1.0`（Up）または `0.0`（Down）
- When: `Tick(dt)`
- Then: 境界扱い（`atBoundary = true`）
- And: 追加の繰越（overflow）がない

#### C-03: 境界を超えた場合 overflow を次ブロックへ繰り越す

- Given: Tick 後 `progress` が `1.0` を超える（Up）/ `0.0` を下回る（Down）
- When: `Tick(dt)`
- Then: 次ブロックへ遷移する（条件を満たす場合）
- And: overflow 分が新しいブロックの `progress` に反映される
- And: 新しい `progress` は `0 < progress < 1` となる（端に張り付かない）

### D. 固定閉塞（Occupancy Invariants）

#### D-01: 1ブロックは同時に1列車まで

- Given: `Block X` を列車 A が占有中
- When: 列車 B が `Block X` に入ろうとする
- Then: 状態遷移は拒否される（列車 B は境界で待機）

#### D-02: 次ブロックが占有されている場合、列車は境界で停止する

- Given: 列車が境界到達し次ブロックへ入ろうとする、次ブロック占有中
- When: `Tick(dt)`
- Then: 列車は現在ブロックの端に張り付く（Up: `progress=1.0` / Down: `progress=0.0`）
- And: `block` は変わらない
- And: 占有は変わらない

#### D-03: 次ブロックが空いたら次 Tick で進入できる

- Given: 前 Tick では占有で停止した列車
- And: 次ブロック占有が解放された
- When: 次の `Tick(dt)`
- Then: 次ブロックに進入する

#### D-04: 占有更新は集中化されている（moveOccupancy）

- Given: ブロック遷移が発生
- When: `Tick(dt)`
- Then: `release(from)` と `occupy(to)` が必ず対で実行される
- And: 遷移後に占有が二重登録されない

注: この項目は「実装構造のテスト」になりやすいので、可能なら black-box（結果で検証）で担保する。

### E. 終端挙動（Terminal Behavior）

仕様を決めた方でテストを固定する。

#### E-01A: 終端停止仕様（Terminal Stop）

- Given: 終端駅に到達する方向へ進行中
- When: 境界に到達する `Tick(dt)`
- Then: 列車は停止状態になる
- And: 次の Tick でも移動しない

#### E-01B: 折返し仕様（Turnback）

- Given: 終端駅に到達する方向へ進行中
- When: 境界に到達する `Tick(dt)`
- Then: 列車は `PendingTurnback`（または同等の状態）になる
- When: 次の `Tick(dt)`
- Then: 方向が反転し、逆方向に進行できる

### F. 列車追加（初期配置）

#### F-01: 同一 `TrainID` の重複追加は禁止

- Given: 既に `TrainID = X` が存在
- When: `TrainID = X` を追加
- Then: エラーになる（`409` 相当でもよい）

#### F-02: 初期ブロックが占有されていたら追加できない

- Given: `Block i` を別列車が占有中
- When: `Block i` に列車を追加
- Then: エラーになる

#### F-03: progress は `0.0..1.0` に制限される

- Given: `progress < 0` または `progress > 1` の入力
- When: 列車を追加
- Then: エラーになる

#### F-04: ブロック ID が路線範囲外なら追加できない

- Given: `BlockID` が存在しない
- When: 追加
- Then: エラーになる

### G. Tick 入力バリデーション

#### G-01: `dt <= 0` は拒否する

- Given: `dt = 0` または負値
- When: `Tick(dt)`
- Then: エラーになる

#### G-02: 大きい `dt` の扱いが仕様化されている

以下どちらかで固定する。

Option A: 複数ブロック跨ぎを許容
- Given: 大きい `dt` で 1 回の Tick で複数ブロックを跨ぐ
- Then: ループ処理で正しく遷移する

Option B: `dt` 上限を設けて拒否
- Given: `dt` が上限超え
- Then: エラーになる

## 出力（DTO）に関するテスト（MVP の安定性のため）

### DTO-01: 駅境界情報が推測ではなく明示される

- Given: 列車が境界にいる
- When: Snapshot DTO を取得
- Then: `atBoundary = true` が含まれる
- And: `stationId` 等が含まれる（必要なら）

### DTO-02: 列車の並び順が安定している

- Given: 複数列車
- When: Snapshot DTO を取得
- Then: `trains[]` の順序が毎回同じ（ID 昇順など）

## 完了条件（Milestone 1 Done）

- [ ] 上記カテゴリ A〜G の主要テストが揃い、CI で安定してパスする
- [ ] 実装がこの仕様からズレる場合は、必ず仕様（本ドキュメント）を更新する
- [ ] シミュレーションのコアが「壊せない」状態になっている
