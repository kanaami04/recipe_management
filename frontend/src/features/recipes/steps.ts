// 作り方(procedure)は API/DB では 1 つの文字列で持ち、UI では手順の配列として扱う。
// 「1 行 = 1 手順」の約束で、改行区切りの文字列 ↔ 手順配列を相互変換する。

// procedure 文字列を手順の配列に分解する。空文字は空配列。
export function splitSteps(procedure: string): string[] {
  return procedure === '' ? [] : procedure.split('\n')
}

// 手順の配列を procedure 文字列に結合する。
export function joinSteps(steps: string[]): string {
  return steps.join('\n')
}
