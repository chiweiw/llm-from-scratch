import { onUnmounted } from "vue";
import { EventsOn } from "../../wailsjs/runtime/runtime";

export function useWailsEvent<T = any>(name: string, handler: (data: T) => void) {
  const off = EventsOn(name, handler as any);
  onUnmounted(off);
}

