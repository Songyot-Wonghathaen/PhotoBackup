import { ref } from 'vue'

// Module-scoped reactive state persists in memory across route transitions
const activeTab = ref<'desc' | 'tag'>('tag')
const searchQuery = ref('')
const selectedTag = ref('')
const selectedTagId = ref<number | null>(null)
const selectedTagExists = ref<boolean | null>(null)
const hasInitialized = ref(false)

export function useGalleryState() {
  function resetFilter() {
    searchQuery.value = ''
    selectedTag.value = ''
    selectedTagId.value = null
    selectedTagExists.value = null
  }

  return {
    activeTab,
    searchQuery,
    selectedTag,
    selectedTagId,
    selectedTagExists,
    hasInitialized,
    resetFilter
  }
}
