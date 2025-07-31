import { renderHook, waitFor } from '@testing-library/react'
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { useFetchMessages } from './useFetchMessages'

// fetch APIをモック
global.fetch = vi.fn()

describe('useFetchMessages', () => {
  beforeEach(() => {
    vi.resetAllMocks()
  })

  it('should fetch message successfully', async () => {
    const mockResponse = {
      message: 'こんにちは、みなさん'
    }

    vi.mocked(fetch).mockResolvedValueOnce({
      ok: true,
      json: async () => mockResponse,
    } as Response)

    const { result } = renderHook(() => useFetchMessages())

    await waitFor(() => {
      expect(result.current.message).toBe('こんにちは、みなさん')
    })

    expect(result.current.loading).toBe(false)
    expect(result.current.error).toBe(null)
    expect(fetch).toHaveBeenCalledWith('http://localhost:8080/message')
  })

  it('should handle fetch error', async () => {
    vi.mocked(fetch).mockRejectedValueOnce(new Error('Network error'))

    const { result } = renderHook(() => useFetchMessages())

    await waitFor(() => {
      expect(result.current.error).toBe('Network error')
    })

    expect(result.current.loading).toBe(false)
    expect(result.current.message).toBe('')
  })
})