<template>
    <div class="tooltip">
        <span class="title">{{ data.Code }}</span>
        <span class="count">{{ data.Count }} 次</span>
    </div>
</template>

<script setup>
defineProps({ data: Object })
</script>

<style scoped>
.key-tooltip-enter-active,
.key-tooltip-leave-active {
    transition: opacity 0.3s ease;
}

.key-tooltip-enter-from,
.key-tooltip-leave-to {
    opacity: 0;
}

.tooltip {
    position: absolute;
    bottom: calc(100% + 8px);
    left: 50%;
    transform: translateX(-50%);
    z-index: 9999;
    pointer-events: none;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    width: max-content;
    white-space: nowrap;
    background: var(--popover);
    color: var(--popover-text);
    padding: 6px 10px;
    border-radius: 8px;
    font-size: 12px;
    font-weight: 400;
    backdrop-filter: blur(14px);
    box-shadow: 0 6px 16px rgba(0, 0, 0, 0.2);
}

.title {
    font-weight: 600;
}

.count {
    opacity: 0.75;
}
</style>
