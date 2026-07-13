<template>
  <figure class="cast">
    <div class="cast__header">
      <div class="cast__traffic" aria-hidden="true">
        <span class="cast__dot cast__dot--close"></span>
        <span class="cast__dot cast__dot--min"></span>
        <span class="cast__dot cast__dot--max"></span>
      </div>
      <div class="cast__title">{{ computedTitle }}</div>
      <div class="cast__controls">
        <button
          class="cast__button"
          type="button"
          :aria-label="playing ? 'Pause recording' : 'Play recording'"
          @click="toggle"
        >
          <svg v-if="playing" viewBox="0 0 24 24" width="15" height="15" fill="currentColor" aria-hidden="true">
            <path d="M6 5h4v14H6V5zm8 0h4v14h-4V5z" />
          </svg>
          <svg v-else viewBox="0 0 24 24" width="15" height="15" fill="currentColor" aria-hidden="true">
            <path d="M8 5v14l11-7L8 5z" />
          </svg>
        </button>
        <button class="cast__button" type="button" aria-label="Restart recording" @click="restart">
          <svg viewBox="0 0 24 24" width="15" height="15" fill="currentColor" aria-hidden="true">
            <path d="M12 5V2L7 7l5 5V8c2.8 0 5 2.2 5 5s-2.2 5-5 5a5 5 0 0 1-4.6-3H4.3a8 8 0 1 0 7.7-10z" />
          </svg>
        </button>
      </div>
    </div>
    <pre ref="screenRef" class="cast__screen" :style="{ height: `${rows * 1.38}em` }"><code>{{ screenText }}</code></pre>
  </figure>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'

type CastEvent = {
  delay: number
  stream: string
  data: string
}

const props = withDefaults(defineProps<{
  src: string
  title?: string
  rows?: number
  autoplay?: boolean
  loop?: boolean
  speed?: number
  maxDelay?: number
}>(), {
  rows: 16,
  autoplay: true,
  loop: true,
  speed: 0.75,
  maxDelay: 1.8,
})

const computedTitle = computed(() => props.title?.trim() || 'Terminal recording')
const screenRef = ref<HTMLElement | null>(null)
const screenText = ref('Loading recording...')
const playing = ref(false)
const events = ref<CastEvent[]>([])
const cursor = ref(0)
const timer = ref<number | null>(null)

function stripTerminalControls(text: string): string {
  return text
    .replace(/\x1b\][^\x07]*(?:\x07|\x1b\\)/g, '')
    .replace(/\x1b\[[0-?]*[ -/]*[@-~]/g, '')
    .replace(/\r\n/g, '\n')
    .replace(/\r/g, '')
}

function appendOutput(data: string) {
  const clean = stripTerminalControls(data)
  if (!clean) return false
  screenText.value += clean
  requestAnimationFrame(() => {
    const el = screenRef.value
    if (el) el.scrollTop = el.scrollHeight
  })
  return true
}

function clearTimer() {
  if (timer.value !== null) {
    window.clearTimeout(timer.value)
    timer.value = null
  }
}

function nextDelay(event: CastEvent, rendered: boolean): number {
  if (!rendered) return 0
  const capped = Math.min(event.delay, props.maxDelay)
  const delay = Math.max(0.35, capped / props.speed)
  return delay * 1000
}

function tick() {
  if (!playing.value) return
  if (cursor.value >= events.value.length) {
    if (props.loop) {
      timer.value = window.setTimeout(() => {
        resetScreen()
        tick()
      }, 1200)
      return
    }
    playing.value = false
    return
  }

  const event = events.value[cursor.value]
  cursor.value += 1
  const rendered = event.stream === 'o' ? appendOutput(event.data) : false
  timer.value = window.setTimeout(tick, nextDelay(event, rendered))
}

function resetScreen() {
  clearTimer()
  screenText.value = ''
  cursor.value = 0
}

function play() {
  if (!events.value.length || playing.value) return
  playing.value = true
  tick()
}

function pause() {
  playing.value = false
  clearTimer()
}

function toggle() {
  if (playing.value) pause()
  else play()
}

function restart() {
  const wasPlaying = playing.value
  playing.value = false
  resetScreen()
  if (wasPlaying || props.autoplay) {
    playing.value = true
    tick()
  }
}

function parseCast(raw: string): CastEvent[] {
  const lines = raw.split(/\r?\n/).filter(Boolean)
  if (lines.length < 2) return []
  return lines.slice(1).map((line) => {
    const parsed = JSON.parse(line) as [number, string, string]
    return {
      delay: Number(parsed[0]) || 0,
      stream: String(parsed[1] || ''),
      data: String(parsed[2] || ''),
    }
  })
}

onMounted(async () => {
  try {
    const response = await fetch(props.src)
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    events.value = parseCast(await response.text())
    resetScreen()
    if (props.autoplay) play()
  } catch {
    screenText.value = `Could not load ${props.src}`
  }
})

onBeforeUnmount(() => {
  pause()
})
</script>

<style scoped>
.cast {
  max-width: 960px;
  margin: 1.25rem auto;
  overflow: hidden;
  border: 1px solid var(--vp-c-divider);
  border-radius: 8px;
  background: #111418;
  box-shadow: 0 8px 22px rgb(0 0 0 / 10%);
  text-align: left;
}

.cast__header {
  display: grid;
  grid-template-columns: auto 1fr auto;
  align-items: center;
  gap: 0.75rem;
  min-height: 2.35rem;
  padding: 0 0.75rem;
  border-bottom: 1px solid rgb(255 255 255 / 8%);
  background: #1b2027;
}

.cast__traffic {
  display: flex;
  gap: 0.42rem;
}

.cast__dot {
  width: 0.72rem;
  height: 0.72rem;
  border-radius: 50%;
}

.cast__dot--close {
  background: #ff5f57;
}

.cast__dot--min {
  background: #ffbd2e;
}

.cast__dot--max {
  background: #28c840;
}

.cast__title {
  overflow: hidden;
  color: #d8dee9;
  font-size: 0.82rem;
  font-weight: 600;
  line-height: 1.2;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.cast__controls {
  display: flex;
  gap: 0.3rem;
}

.cast__button {
  display: grid;
  width: 1.65rem;
  height: 1.65rem;
  place-items: center;
  border: 0;
  border-radius: 6px;
  color: #d8dee9;
  background: rgb(255 255 255 / 7%);
  cursor: pointer;
}

.cast__button:hover {
  background: rgb(255 255 255 / 14%);
}

.cast__screen {
  display: block;
  box-sizing: border-box;
  width: 100%;
  max-height: 24rem;
  margin: 0;
  overflow: auto;
  padding: 0.85rem 1rem;
  color: #e7edf4;
  background: #111418;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
  font-size: 0.78rem;
  line-height: 1.32;
  text-align: left;
  tab-size: 8;
  white-space: pre;
}

.cast__screen code {
  display: block;
  width: max-content;
  min-width: 100%;
  color: inherit;
  background: transparent;
  text-align: left;
  white-space: pre;
}

@media (max-width: 640px) {
  .cast__header {
    padding: 0 0.55rem;
  }

  .cast__screen {
    padding: 0.8rem;
    font-size: 0.78rem;
  }
}
</style>
