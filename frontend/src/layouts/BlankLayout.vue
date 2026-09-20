<template>
  <div class="blank-layout sd-grid-bg">
    <div class="blank-layout__glow blank-layout__glow--cyan" />
    <div class="blank-layout__glow blank-layout__glow--purple" />
    <div class="blank-layout__inner">
      <router-view />
    </div>
  </div>
</template>

<script setup lang="ts"></script>

<style scoped lang="scss">
.blank-layout {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  padding: 40px 16px;
  overflow: hidden;
}

.blank-layout__inner {
  position: relative;
  z-index: 2;
  width: 100%;
  max-width: 460px;
}

.blank-layout__glow {
  position: absolute;
  width: 460px;
  height: 460px;
  border-radius: 50%;
  filter: blur(120px);
  opacity: 0.35;
}

.blank-layout__glow--cyan {
  top: -140px;
  left: -120px;
  background: #22d3ee;
}

.blank-layout__glow--purple {
  right: -140px;
  bottom: -160px;
  background: #a855f7;
}
</style>
