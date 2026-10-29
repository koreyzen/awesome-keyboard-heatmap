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
  display: grid;
  grid-template-rows: 1fr auto 1fr;
  justify-items: center;
  background: #f2f3f5;
  width: 100%;
  height: 100vh;
}
</style>
