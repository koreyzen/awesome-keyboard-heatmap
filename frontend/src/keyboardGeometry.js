// Use standard key units for every row; pixel dimensions follow the container.
export function createKeyboardGeometry(rows) {
  const columns = Math.max(...rows.map(row => row.reduce((sum, key) => sum + key.W, 0)))
  const keys = rows.flatMap((row, rowIndex) => {
    let x = 0
    return row.flatMap(column => {
      const stack = column.Stack || [column]
      let y = rowIndex + (column.Stack ? 0 : Math.max(0, 1 - (column.H || 1)))
      const positioned = stack.map(key => {
        const height = key.H || 1 / stack.length
        const result = { ...key, X: x, Y: y, W: column.W, H: height }
        y += height
        return result
      })
      x += column.W
      return positioned.filter(key => key.Code !== 'None')
    })
  })
  return { keys, columns, rows: rows.length }
}
