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
  box-sizing: border-box;
  padding: 50px 24px 24px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  background: #f2f3f5;
  width: 100%;
  height: 100vh;
}
</style>
