<script setup>
import Keyboard from "./components/Keyboard.vue";
</script>

<template>
  <div class="keyboard-wrapper">
    <Keyboard />
  </div>
</template>

<style scoped>
.keyboard-wrapper {
  --toolbar-height: 52px;
  box-sizing: border-box;
  padding: 0 24px;
  display: grid;
  grid-template-rows: var(--toolbar-height) minmax(0, 1fr) var(--toolbar-height);
  justify-items: center;
  background: #f2f3f5;
  width: 100%;
  height: 100vh;
}
</style>
