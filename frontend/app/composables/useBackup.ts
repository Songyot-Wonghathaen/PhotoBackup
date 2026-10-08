import { ref } from 'vue'

export function useBackup() {
  const sourcePath = ref('/Volumes/USBDrive/DCIM')
  const destPath = ref('~/Pictures/PhotoBackup')
  const isWailsAvailable = ref(false)

  const checkWails = () => {
    isWailsAvailable.value = typeof (window as any)?.go?.main?.App !== 'undefined'
    return isWailsAvailable.value
  }

  return {
    sourcePath,
    destPath,
    isWailsAvailable,
    checkWails
  }
}
