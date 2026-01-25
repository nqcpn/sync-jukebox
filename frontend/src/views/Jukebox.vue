<template>
  <div class="jukebox-layout">
    <header>
      <div class="header-brand">
        <img src="@/assets/icon.png" alt="SyncJukebox Icon" class="icon" />
        <div class="title">
          SyncJukebox
        </div>
      </div>

      <!-- 修改: 将右侧控件分组 -->
      <div class="header-controls">
        <OnlineUsers /> <!-- 新增: 在线用户组件 -->
        <button @click="handleLogout" class="logout-button">Logout</button>
      </div>
    </header>
    <main>
      <div class="left-panel">
        <MediaLibrary />
      </div>
      <div class="right-panel">
        <Playlist />
      </div>
    </main>
    <footer>
      <PlayerControls />
    </footer>

    <PlaybackPermissionModal :visible="showPermissionModal" @confirm="handlePermissionConfirm" />
  </div>
</template>

<script setup>
import { ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import MediaLibrary from '../components/MediaLibrary.vue';
import Playlist from '../components/Playlist.vue';
import PlayerControls from '../components/PlayerControls.vue';
import PlaybackPermissionModal from '../components/PlaybackPermissionModal.vue';
import OnlineUsers from '../components/OnlineUsers.vue'; // <-- 新增: 导入组件
import { usePlayerStore } from '@/stores/player';

const store = usePlayerStore();
const router = useRouter();
const showPermissionModal = ref(false);

const handleLogout = () => {
  store.logout();
  router.push({ name: 'login' });
};

const handlePermissionConfirm = async () => {
  try {
    await store.play();
    showPermissionModal.value = false;
  } catch (e) {
    console.error("依然无法播放:", e);
  }
};

watch(() => store.playbackError, (newError) => {
  if (newError && newError.name === 'NotAllowedError') {
    showPermissionModal.value = true;
  }
});
</script>

<style scoped>
.jukebox-layout {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.icon {
  width: 32px;
  height: 32px;
  margin-right: 0.5rem;
}

header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.5rem;
  background-color: #181818;
  font-size: 1.2rem;
  font-weight: bold;
  flex-shrink: 0;
}

.header-brand {
  display: flex;
  align-items: center;
}

/* 新增: 右侧控件容器样式 */
.header-controls {
  display: flex;
  align-items: center;
  gap: 1rem; /* 为控件之间添加间距 */
}

.logout-button {
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

.logout-button:hover {
  background-color: #fff;
  color: #121212;
}

main {
  display: flex;
  flex: 1;
  overflow: hidden;
  padding: 1.5rem;
  gap: 1.5rem;
}

.left-panel,
.right-panel {
  flex: 1;
  background-color: #181818;
  border-radius: 8px;
  padding: 1.5rem;
  min-width: 0;
}

footer {
  background-color: #181818;
  flex-shrink: 0;
}
</style>
