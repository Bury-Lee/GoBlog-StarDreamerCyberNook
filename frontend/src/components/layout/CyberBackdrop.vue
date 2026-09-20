<template>
  <div class="cyber-backdrop">
    <div class="cyber-backdrop__grid" />
    <div class="cyber-backdrop__scan" />
  </div>
</template>

<script setup lang="ts"></script>

<style scoped lang="scss">
.cyber-backdrop {
  position: fixed;
  inset: 0;
  z-index: 0;
  pointer-events: none;
  overflow: hidden;
}

.cyber-backdrop__grid {
  position: absolute;
  inset: -20%;
  background-image:
    linear-gradient(rgba(34, 211, 238, 0.06) 1px, transparent 1px),
    linear-gradient(90deg, rgba(34, 211, 238, 0.06) 1px, transparent 1px);
  background-size: 56px 56px;
  transform: perspective(600px) rotateX(58deg) translateY(-10%);
  mask-image: linear-gradient(180deg, #000 0%, transparent 72%);
}

.cyber-backdrop__scan {
  position: absolute;
  inset: 0;
  background: repeating-linear-gradient(
    180deg,
    rgba(148, 163, 184, 0.03) 0px,
    rgba(148, 163, 184, 0.03) 1px,
    transparent 1px,
    transparent 4px
  );
  opacity: 0.6;
}
</style>
