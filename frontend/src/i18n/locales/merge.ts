type LocaleMessages = Record<string, unknown>

function isMessageObject(value: unknown): value is LocaleMessages {
  return value !== null && typeof value === 'object' && !Array.isArray(value)
}

export function mergeLocaleAdditions<T extends LocaleMessages>(base: T, additions: LocaleMessages): T {
  const result: LocaleMessages = { ...base }

  for (const [key, value] of Object.entries(additions)) {
    if (!(key in result)) {
      result[key] = value
    } else if (isMessageObject(result[key]) && isMessageObject(value)) {
      result[key] = mergeLocaleAdditions(result[key], value)
    }
  }

  return result as T
}

export function mergeLocaleOverrides<T extends LocaleMessages>(base: T, overrides: LocaleMessages): T {
  const result: LocaleMessages = { ...base }

  for (const [key, value] of Object.entries(overrides)) {
    if (isMessageObject(result[key]) && isMessageObject(value)) {
      result[key] = mergeLocaleOverrides(result[key], value)
    } else {
      result[key] = value
    }
  }

  return result as T
}
