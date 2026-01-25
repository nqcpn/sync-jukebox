<template>
  <!-- 1. 添加 ref 以便在脚本中引用此元素 -->
  <div class="online-users-widget" ref="widgetRef">
    <button class="widget-button" @click="toggleDropdown">
      <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-people-fill" viewBox="0 0 16 16">
        <path d="M7 14s-1 0-1-1 1-4 5-4 5 3 5 4-1 1-1 1zm4-6a3 3 0 1 0 0-6 3 3 0 0 0 0 6m-5.784 6A2.238 2.238 0 0 1 5 13c0-1.355.68-2.75 1.936-3.72A6.325 6.325 0 0 0 5 9c-4 0-5 3-5 4s1 1 1 1zM4.5 8a2.5 2.5 0 1 0 0-5 2.5 2.5 0 0 0 0 5"/>
      </svg>
      <span>{{ onlineUserCount }}</span>
    </button>

    <!-- 2. 将模态框结构改为下拉菜单结构 -->
    <div v-if="isDropdownVisible" class="dropdown-menu">
      <div class="dropdown-header">Online Users ({{ onlineUserCount }})</div>
      <ul v-if="onlineUsers.length > 0" class="user-list">
        <li v-for="user in onlineUsers" :key="user.id">{{ user.username }}</li>
      </ul>
      <div v-else class="empty-message">No other users online.</div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onUnmounted } from 'vue';
import { usePlayerStore } from '@/stores/player';

const store = usePlayerStore();
const isDropdownVisible = ref(false);
const widgetRef = ref(null); // 用于引用组件的根DOM元素

const onlineUserCount = computed(() => store.onlineUserCount);
const onlineUsers = computed(() => store.onlineUsers);

const toggleDropdown = () => {
  isDropdownVisible.value = !isDropdownVisible.value;
};

const closeDropdown = () => {
  isDropdownVisible.value = false;
};

// 3. 实现 "点击外部区域关闭" 的逻辑
const handleClickOutside = (event) => {
  // 如果下拉列表是可见的，并且点击的目标不在组件内部
  if (isDropdownVisible.value && widgetRef.value && !widgetRef.value.contains(event.target)) {
    closeDropdown();
  }
};

// 4. 使用 watch 监听下拉列表的显示状态，动态添加/移除事件监听器
watch(isDropdownVisible, (isVisible) => {
  if (isVisible) {
    // 延迟到下一个tick添加，以防止切换按钮的点击事件立即触发 'handleClickOutside'
    setTimeout(() => {
      document.addEventListener('click', handleClickOutside);
    }, 0);
  } else {
    document.removeEventListener('click', handleClickOutside);
  }
});

// 5. 在组件卸载时，确保移除事件监听器，防止内存泄漏
onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside);
});
</script>

<style scoped>
/* 关键: 父容器需要相对定位 */
.online-users-widget {
  position: relative;
}

.widget-button {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  font-size: 0.9rem;
  font-weight: bold;
  color: #fff;
  background-color: transparent;
  border: 1px solid #535353;
  border-radius: 20px;
  cursor: pointer;
  transition: background-color 0.2s, color 0.2s;
}

.widget-button:hover {
  background-color: #282828;
}

/* 新的下拉菜单样式 */
.dropdown-menu {
  position: absolute;
  top: calc(100% + 8px); /* 在按钮下方，并留出8px间距 */
  right: 0;
  background-color: #282828;
  border-radius: 6px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.5);
  min-width: 220px;
  z-index: 1000;
  overflow: hidden; /* 配合border-radius */
}

.dropdown-header {
  padding: 0.75rem 1rem;
  font-size: 0.9rem;
  font-weight: bold;
  border-bottom: 1px solid #404040;
}

.user-list {
  list-style: none;
  padding: 0;
  margin: 0;
  max-height: 300px;
  overflow-y: auto;
}

.user-list li {
  padding: 0.6rem 1rem;
  font-size: 0.85rem;
  white-space: nowrap;
}

.user-list li:not(:last-child) {
  border-bottom: 1px solid #333;
}

.user-list li:hover {
  background-color: #3e3e3e;
}

.empty-message {
  padding: 1rem;
  color: #b3b3b3;
  font-size: 0.85rem;
  text-align: center;
}
</style>
