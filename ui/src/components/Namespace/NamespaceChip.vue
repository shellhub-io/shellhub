<template>
  <div class="namespace-icon">
    <svg ref="svgRef" :width="size" :height="size" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from "vue";
import { toSvg, configure } from "jdenticon";
import { useTheme } from "vuetify";
import convert from "color-convert";

interface Props {
  name?: string;
  size?: number;
}

const props = withDefaults(defineProps<Props>(), {
  name: "",
  size: 32,
});

const theme = useTheme();
const svgRef = ref<SVGElement | null>(null);

// Get hue from primary color
const getPrimaryHue = (): number => {
  const primaryColor = theme.current.value.colors.primary;
  const hex = primaryColor.replace("#", "");
  const [hue] = convert.hex.hsl(hex);
  return hue;
};

const primaryHue = getPrimaryHue();

// Configure Jdenticon with primary color variations
configure({
  hues: [primaryHue],
  lightness: {
    color: [0.40, 0.60],
    grayscale: [0.30, 0.70],
  },
  saturation: {
    color: 0.50,
    grayscale: 0.00,
  },
  backColor: "#00000000",
});

const updateIcon = () => {
  if (svgRef.value && props.name) {
    svgRef.value.innerHTML = toSvg(props.name, props.size);
  }
};

onMounted(() => {
  updateIcon();
});

watch(() => props.name, () => {
  updateIcon();
});
</script>

<style scoped>
.namespace-icon {
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
