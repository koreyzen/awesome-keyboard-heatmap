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
  --page-inset: 19px;
  box-sizing: border-box;
  padding: 0 var(--page-inset);
  display: grid;
  grid-template-rows: var(--toolbar-height) minmax(0, 1fr) var(--toolbar-height);
  justify-items: center;
  background: var(--page);
  width: 100%;
  height: 100vh;
}
</style>
