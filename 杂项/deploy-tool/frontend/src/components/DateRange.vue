<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import type { PropType } from "vue";
import { X } from "lucide-vue-next";

type Range = { start: string; end: string };
const props = defineProps({
  modelValue: {
    type: Object as PropType<Range>,
    required: true,
  },
});
const emit = defineEmits<{
  (e: "update:modelValue", value: Range): void;
}>();

const startDate = ref<Date | null>(null);
const endDate = ref<Date | null>(null);

function parseYMD(s: string): Date | null {
  if (!s) return null;
  const parts = s.split("-");
  if (parts.length !== 3) return null;
  const y = Number(parts[0]);
  const m = Number(parts[1]);
  const d = Number(parts[2]);
  const dt = new Date(y, m - 1, d);
  if (isNaN(dt.getTime())) return null;
  return dt;
}

function formatYMD(d: Date): string {
  const y = d.getFullYear();
  const m = d.getMonth() + 1;
  const dd = d.getDate();
  const mmStr = m < 10 ? "0" + m : "" + m;
  const ddStr = dd < 10 ? "0" + dd : "" + dd;
  return `${y}-${mmStr}-${ddStr}`;
}

// 左侧日历的月份基准（右侧 = 左侧 +1 个月）
const leftMonth = ref(
  new Date(new Date().getFullYear(), new Date().getMonth() - 1, 1)
);

const rightMonth = computed(
  () =>
    new Date(leftMonth.value.getFullYear(), leftMonth.value.getMonth() + 1, 1)
);

watch(
  () => props.modelValue,
  (val) => {
    startDate.value = parseYMD(val.start);
    endDate.value = parseYMD(val.end);
    const base = startDate.value || endDate.value;
    if (base) {
      // 让起始日落在左侧日历
      leftMonth.value = new Date(base.getFullYear(), base.getMonth() - 1, 1);
    }
  },
  { deep: true, immediate: true }
);

const open = ref(false);
const root = ref<HTMLElement | null>(null);
function onDocClick(e: MouseEvent) {
  if (!open.value) return;
  const t = e.target as Node | null;
  if (root.value && t && !root.value.contains(t)) {
    open.value = false;
  }
}
onMounted(() => document.addEventListener("click", onDocClick));
onBeforeUnmount(() => document.removeEventListener("click", onDocClick));

const hasValue = computed(
  () => !!(props.modelValue.start || props.modelValue.end)
);

const label = computed(() => {
  if (props.modelValue.start && props.modelValue.end)
    return `${props.modelValue.start} — ${props.modelValue.end}`;
  if (props.modelValue.start) return `${props.modelValue.start} —`;
  if (props.modelValue.end) return `— ${props.modelValue.end}`;
  return "选择日期范围";
});

function clearRange(e?: MouseEvent) {
  if (e) e.stopPropagation();
  startDate.value = null;
  endDate.value = null;
  emit("update:modelValue", { start: "", end: "" });
  open.value = false;
}

function isSameDay(a: Date, b: Date): boolean {
  return (
    a.getFullYear() === b.getFullYear() &&
    a.getMonth() === b.getMonth() &&
    a.getDate() === b.getDate()
  );
}

function isBetween(day: Date, start: Date, end: Date): boolean {
  const t = day.getTime();
  return t >= start.getTime() && t <= end.getTime();
}

function selectDay(day: Date) {
  if (!startDate.value || (startDate.value && endDate.value)) {
    startDate.value = day;
    endDate.value = null;
    emit("update:modelValue", { start: formatYMD(day), end: "" });
    return;
  }
  if (startDate.value && !endDate.value) {
    if (day.getTime() < startDate.value.getTime()) {
      endDate.value = startDate.value;
      startDate.value = day;
    } else {
      endDate.value = day;
    }
    emit("update:modelValue", {
      start: formatYMD(startDate.value),
      end: formatYMD(endDate.value!),
    });
    open.value = false;
  }
}

function prevMonth() {
  leftMonth.value = new Date(
    leftMonth.value.getFullYear(),
    leftMonth.value.getMonth() - 1,
    1
  );
}

function nextMonth() {
  leftMonth.value = new Date(
    leftMonth.value.getFullYear(),
    leftMonth.value.getMonth() + 1,
    1
  );
}

const weekLabels = ["日", "一", "二", "三", "四", "五", "六"];

function buildDays(base: Date): { date: Date; inMonth: boolean }[] {
  const first = new Date(base.getFullYear(), base.getMonth(), 1);
  const start = new Date(first);
  start.setDate(1 - start.getDay());
  const items: { date: Date; inMonth: boolean }[] = [];
  for (let i = 0; i < 42; i += 1) {
    const d = new Date(start);
    d.setDate(start.getDate() + i);
    items.push({
      date: d,
      inMonth: d.getMonth() === base.getMonth(),
    });
  }
  return items;
}

const leftDays = computed(() => buildDays(leftMonth.value));
const rightDays = computed(() => buildDays(rightMonth.value));

function dayClass(day: Date, inMonth: boolean): string {
  const classes = [
    "h-8",
    "w-8",
    "rounded",
    "text-sm",
    "flex",
    "items-center",
    "justify-center",
    "hover:bg-accent",
    "cursor-pointer",
  ];
  if (!inMonth) classes.push("text-muted-foreground/40");
  const s = startDate.value;
  const e = endDate.value;
  if (s && e && isBetween(day, s, e)) classes.push("bg-primary/10");
  if (s && isSameDay(day, s))
    classes.push("bg-primary", "text-primary-foreground");
  if (e && isSameDay(day, e))
    classes.push("bg-primary", "text-primary-foreground");
  const now = new Date();
  if (isSameDay(day, now)) classes.push("border", "border-primary/50");
  return classes.join(" ");
}
</script>

<template>
  <div ref="root" class="relative inline-block w-full">
    <!-- 触发器 -->
    <div
      class="flex items-center gap-2 rounded border px-3 py-2 w-full cursor-pointer bg-background"
      @click.stop="open = !open"
    >
      <div
        class="flex-1 text-sm"
        :class="label === '选择日期范围' ? 'text-muted-foreground' : ''"
      >
        {{ label }}
      </div>
      <!-- X 清除按钮：只在有选中值时显示 -->
      <button
        v-if="hasValue"
        class="text-muted-foreground hover:text-foreground transition-colors"
        @click.stop="clearRange"
        aria-label="清除"
      >
        <X class="w-4 h-4" />
      </button>
    </div>

    <!-- 双月弹出日历 -->
    <div
      v-if="open"
      class="absolute z-50 mt-2 rounded border bg-popover shadow-lg p-4"
      style="min-width: 580px; left: auto; right: 0"
    >
      <div class="flex gap-6">
        <!-- 左侧月份 -->
        <div class="flex-1">
          <div class="flex items-center justify-between mb-2">
            <button
              class="rounded border px-2 py-1 text-sm hover:bg-accent"
              @click="prevMonth"
            >
              ‹
            </button>
            <div class="text-sm font-medium">
              {{ leftMonth.getFullYear() }}年{{ leftMonth.getMonth() + 1 }}月
            </div>
            <div class="w-8" />
            <!-- 占位，保持标题居中 -->
          </div>
          <div
            class="grid grid-cols-7 gap-1 mb-1 text-xs text-muted-foreground"
          >
            <div v-for="w in weekLabels" :key="w" class="text-center">
              {{ w }}
            </div>
          </div>
          <div class="grid grid-cols-7 gap-1">
            <button
              v-for="item in leftDays"
              :key="item.date.toISOString()"
              type="button"
              :class="dayClass(item.date, item.inMonth)"
              @click="selectDay(item.date)"
            >
              {{ item.date.getDate() }}
            </button>
          </div>
        </div>

        <!-- 分隔线 -->
        <div class="w-px bg-border self-stretch" />

        <!-- 右侧月份 -->
        <div class="flex-1">
          <div class="flex items-center justify-between mb-2">
            <div class="w-8" />
            <!-- 占位 -->
            <div class="text-sm font-medium">
              {{ rightMonth.getFullYear() }}年{{ rightMonth.getMonth() + 1 }}月
            </div>
            <button
              class="rounded border px-2 py-1 text-sm hover:bg-accent"
              @click="nextMonth"
            >
              ›
            </button>
          </div>
          <div
            class="grid grid-cols-7 gap-1 mb-1 text-xs text-muted-foreground"
          >
            <div v-for="w in weekLabels" :key="w" class="text-center">
              {{ w }}
            </div>
          </div>
          <div class="grid grid-cols-7 gap-1">
            <button
              v-for="item in rightDays"
              :key="item.date.toISOString()"
              type="button"
              :class="dayClass(item.date, item.inMonth)"
              @click="selectDay(item.date)"
            >
              {{ item.date.getDate() }}
            </button>
          </div>
        </div>
      </div>

      <!-- 底部清除按钮 -->
      <div class="flex justify-end mt-3 pt-3 border-t">
        <button
          class="rounded border px-3 py-1 text-sm hover:bg-accent"
          @click="clearRange"
        >
          清除
        </button>
      </div>
    </div>
  </div>
</template>
