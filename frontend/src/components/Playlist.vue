<template>
  <div class="playlist-container">
    <div class="playlist-header">
      <h2>Playlist</h2>
      <!-- *** 新增: 按钮容器 *** -->
      <div class="header-actions">
        <!-- *** 新增: 定位按钮 *** -->
        <button
            class="locate-btn"
            @click="scrollToCurrentSong"
            title="Locate Current Song"
            :disabled="!store.currentSongId"
        >
          <IconLocate/>
        </button>

        <button
            class="shuffle-btn"
            @click="handleShuffle"
            title="Shuffle Playlist"
            :disabled="!store.playlist || store.playlist.length < 2"
        >
          <IconShuffle/>
        </button>
      </div>
    </div>
    <ul v-if="store.playlist && store.playlist.length > 0" class="song-list">
      <!--
        *** 修改: 添加 :ref 来捕获每个列表项的 DOM 元素 ***
        我们使用一个函数 ref 来将每个 li 元素存入一个数组中。
      -->
      <li v-for="(item, index) in store.playlist" :key="item.song_id"
          :ref="el => { if (el) songItemRefs[index] = el }" class="song-item" :class="{
        'is-playing': store.currentSongId === item.song_id,
        'is-dragging': draggedItemIndex === index,
        'drag-over-top': dragOverIndex === index && dragOverPosition === 'top',
        'drag-over-bottom': dragOverIndex === index && dragOverPosition === 'bottom'
      }" draggable="true" @dblclick="handleDoubleClick(item)" @dragstart="onDragStart($event, index, item)"
          @dragover.prevent="onDragOver($event, index)" @dragend="onDragEnd" @drop="onDrop($event, index)">
        <div class="song-details">
          <span class="song-title">{{ item.song.title }}</span>
          <span class="song-artist">{{ item.song.artist }}</span>
        </div>

        <div class="song-actions">
          <button
              v-if="store.currentSongId !== item.song_id"
              @click.stop="store.addSongNextInPlaylist(item.song.id)"
              title="Play Next"
              class="play-next-btn"
          >
            +⤋
          </button>

          <button class="remove-btn" @click.stop="handleRemove(item)" title="Remove from playlist">
            <IconTrash></IconTrash>
          </button>
        </div>
      </li>
    </ul>
    <p v-else class="empty-message">The playlist is currently empty.</p>
  </div>
</template>

<script setup>
// *** 修改: 引入 onBeforeUpdate 用于维护 refs 数组 ***
import {ref, onBeforeUpdate} from 'vue';
import IconTrash from '@/components/icons/IconTrash.vue';
import IconShuffle from '@/components/icons/IconShuffle.vue';
import {usePlayerStore} from '@/stores/player';
import IconLocate from "@/components/icons/IconLocate.vue";

const store = usePlayerStore();

// --- 新增: 定位到当前歌曲的逻辑 ---
// 创建一个 ref 来存储所有歌曲 li 元素的 DOM 引用
const songItemRefs = ref([]);

// 在每次组件更新前，清空 refs 数组，以防止因列表项重新排序或删除而导致的引用错乱
onBeforeUpdate(() => {
  songItemRefs.value = [];
});

// 定位按钮的点击处理函数
const scrollToCurrentSong = () => {
  if (!store.currentSongId) return;

  // 1. 在播放列表中找到当前歌曲的索引
  const currentIndex = store.playlist.findIndex(
      (item) => item.song_id === store.currentSongId
  );

  if (currentIndex > -1) {
    // 2. 从 refs 数组中获取对应的 DOM 元素
    const targetElement = songItemRefs.value[currentIndex];

    // 3. 如果元素存在，则滚动到视图中央
    if (targetElement) {
      targetElement.scrollIntoView({
        behavior: 'smooth', // 平滑滚动
        block: 'center',    // 垂直方向上居中
      });
    }
  }
};


// --- 处理打乱点击 ---
const handleShuffle = () => {
  store.shufflePlaylist();
};

const handleRemove = (item) => {
  store.removeSongFromPlaylist(item.song_id);
};

// --- 双击播放功能 ---
const handleDoubleClick = (item) => {
  store.playSpecificSong(item.song_id);
};

// --- 拖拽排序逻辑 (保持不变) ---
const draggedItemIndex = ref(null);
const draggedItemData = ref(null);
const dragOverIndex = ref(null);
const dragOverPosition = ref(null);

const onDragStart = (event, index, item) => {
  draggedItemIndex.value = index;
  draggedItemData.value = item;
  event.dataTransfer.effectAllowed = 'move';
  event.dataTransfer.setData('text/plain', index);
};

const onDragEnd = () => {
  draggedItemIndex.value = null;
  draggedItemData.value = null;
  dragOverIndex.value = null;
  dragOverPosition.value = null;
};

const onDragOver = (event, index) => {
  event.preventDefault();
  if (draggedItemIndex.value !== null && draggedItemIndex.value !== index) {
    dragOverIndex.value = index;
    const targetRect = event.currentTarget.getBoundingClientRect();
    const hoverY = event.clientY - targetRect.top;
    const threshold = targetRect.height / 2;
    dragOverPosition.value = hoverY < threshold ? 'top' : 'bottom';
  }
};

const onDrop = (event, targetIndex) => {
  const sourceIndex = draggedItemIndex.value;
  const item = draggedItemData.value;
  if (sourceIndex === null || sourceIndex === targetIndex) {
    return;
  }

  let finalIndex = targetIndex;
  if (dragOverPosition.value === 'bottom') {
    finalIndex++;
  }
  if (sourceIndex < targetIndex) {
    finalIndex--;
  }

  store.movePlaylistItem(item.song_id, finalIndex);
};
</script>

<style scoped>
/* 样式无需修改，保持原样... */
.playlist-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-shrink: 0;
  margin-bottom: 1rem;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid #282828;
}

.playlist-header h2 {
  margin: 0;
  padding: 0;
  border: none;
}

/* *** 新增: 头部操作按钮容器样式 *** */
.header-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem; /* 给按钮之间添加一些间距 */
}

/* *** 新增: 定位按钮样式 (复用 shuffle-btn) *** */
.shuffle-btn,
.locate-btn {
  background: transparent;
  border: none;
  cursor: pointer;
  color: #b3b3b3;
  width: 32px;
  height: 32px;
  padding: 0;
  border-radius: 50%;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* *** 修改: 将 :hover 等伪类应用于两个按钮 *** */
.shuffle-btn:hover,
.locate-btn:hover {
  color: #fff;
  background-color: rgba(255, 255, 255, 0.1);
  transform: scale(1.05);
}

.shuffle-btn:active,
.locate-btn:active {
  color: #1db954;
  transform: scale(0.95);
  background-color: rgba(255, 255, 255, 0.15);
}

.shuffle-btn:disabled,
.locate-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
  transform: none;
  color: #b3b3b3;
}


.playlist-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.song-list,
.empty-message {
  flex: 1;
  overflow-y: auto;
}

.song-list {
  list-style: none;
  padding-right: 0.1rem;
  margin: 0;
}

.song-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 0.5rem;
  padding-right: 1.0rem;
  border-top: 2px solid transparent;
  border-bottom: 2px solid transparent;
  box-shadow: 0 1px 0 #282828;
  margin-bottom: 1px;
  transition: background-color 0.2s ease, border-top 0.1s;
  cursor: pointer;
  position: relative;
  user-select: none;
}

.song-item:hover {
  background-color: #2a2a2a;
}

.song-item:hover .song-actions {
  opacity: 1;
  transform: translateX(0);
}

.song-item[draggable="true"] {
  cursor: grab;
}

.song-item[draggable="true"]:active {
  cursor: grabbing;
}

.song-item.drag-over-top {
  border-top-color: #1db954;
  background-color: #333;
}

.song-item.drag-over-bottom {
  border-bottom-color: #1db954;
  background-color: #333;
  box-shadow: none;
}

.song-item.is-dragging {
  opacity: 0.5;
  background-color: #2a2a2a;
}

.song-item:last-child {
  border-bottom: none;
}

.song-details {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
  flex: 1;
  min-width: 0;
  pointer-events: none;
}

.song-title {
  font-weight: 500;
  color: #fff;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.song-artist {
  font-size: 0.8rem;
  color: #b3b3b3;
}

.song-item.is-playing .song-title,
.song-item.is-playing .song-artist {
  color: #1db954;
  font-weight: bold;
}

.song-actions {
  display: flex;
  flex-shrink: 0;
  gap: 0.5rem;
  opacity: 0;
  transform: translateX(10px);
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.song-actions .play-next-btn {
  background: none;
  border: 1px solid #6a6a6a;
  color: #fff;
  border-radius: 50%;
  width: 28px;
  height: 28px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0;
  font-size: 1rem;
  font-weight: bold;
  line-height: 1;
}

.song-actions .play-next-btn:hover {
  border-color: #fff;
  transform: scale(1.05);
}

.remove-btn {
  background: transparent;
  border: none;
  cursor: pointer;
  padding: 0;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #b3b3b3;
  transition: all 0.2s ease;
}

.remove-btn:hover {
  background-color: rgba(255, 255, 255, 0.1);
  color: #ff5555;
}

.remove-btn:active {
  transform: scale(0.95);
}

.empty-message {
  color: #b3b3b3;
  text-align: center;
  margin-top: 2rem;
  font-style: italic;
}

.song-list::-webkit-scrollbar {
  width: 8px;
}

.song-list::-webkit-scrollbar-track {
  background: #181818;
  border-radius: 10px;
}

.song-list::-webkit-scrollbar-thumb {
  background-color: #535353;
  border-radius: 10px;
  border: 2px solid #181818;
}

.song-list::-webkit-scrollbar-thumb:hover {
  background-color: #777;
}
</style>
