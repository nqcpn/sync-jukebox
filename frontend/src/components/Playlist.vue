<template>
  <div class="playlist-container">
    <div class="playlist-header">
      <h2>Playlist</h2>
      <button
          class="shuffle-btn"
          @click="handleShuffle"
          title="Shuffle Playlist"
          :disabled="!store.playlist || store.playlist.length < 2"
      >
        <IconShuffle />
      </button>
    </div>
    <ul v-if="store.playlist && store.playlist.length > 0" class="song-list">
      <li v-for="(item, index) in store.playlist" :key="item.song_id" class="song-item" :class="{
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

        <!--
          *** 新增：歌曲操作按钮容器 ***
          1. 使用 <div> 包裹所有按钮，方便统一样式和 hover 效果。
          2. 应用 .song-actions class 来实现悬停时整体显示。
         -->
        <div class="song-actions">
          <!--
            *** 新增：“播放下一首”按钮 ***
            1. v-if: 仅当歌曲不是当前播放的歌曲时显示。
            2. @click.stop: 调用 handlePlayNext 方法，并阻止事件冒泡。
            3. class & title: 复用 medialib 中的样式和提示。
           -->
          <button
              v-if="store.currentSongId !== item.song_id"
              @click.stop="handlePlayNext(item)"
              title="Play Next"
              class="play-next-btn"
          >
            +⤋
          </button>

          <!--
            *** 修改：移除按钮 ***
            1. 移入 .song-actions 容器中。
            2. @click.stop: 确保点击删除时不会触发双击播放等父级事件。
           -->
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
import { ref } from 'vue';
import IconTrash from '@/components/icons/IconTrash.vue';
import IconShuffle from '@/components/icons/IconShuffle.vue';
import { usePlayerStore } from '@/stores/player';
const store = usePlayerStore();

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

// *** 新增：处理“播放下一首”点击事件 ***
const handlePlayNext = (itemToMove) => {
  // 如果播放列表少于2首歌，无需操作
  if (!store.playlist || store.playlist.length < 2) return;

  // 找到当前播放歌曲的索引
  const currentPlayingIndex = store.currentSongId
      ? store.playlist.findIndex(p => p.song_id === store.currentSongId)
      : -1; // 如果没有歌曲在播放，返回-1

  // 目标位置是当前播放歌曲的下一位。如果没有歌曲播放，则移到列表顶部（位置0）。
  const targetPosition = currentPlayingIndex + 1;

  // 找到要移动的歌曲的当前索引，用于处理拖拽逻辑中的偏移量
  const sourcePosition = store.playlist.findIndex(p => p.song_id === itemToMove.song_id);

  // 如果歌曲已经就在目标位置，则什么也不做
  if (sourcePosition === targetPosition) {
    return;
  }

  // 调用 Pinia store 中已有的移动方法，这个方法应该已经处理了索引修正的逻辑
  // 和拖拽释放(onDrop)时调用的是同一个 action，复用后端交互逻辑
  store.movePlaylistItem(itemToMove.song_id, targetPosition);
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
/* playlist-header 和 shuffle-btn 样式保持不变 */
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

.shuffle-btn {
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
.shuffle-btn:hover {
  color: #fff;
  background-color: rgba(255, 255, 255, 0.1);
  transform: scale(1.05);
}
.shuffle-btn:active {
  color: #1db954;
  transform: scale(0.95);
  background-color: rgba(255, 255, 255, 0.15);
}
.shuffle-btn:disabled {
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

/* *** 修改：让按钮容器在悬停时显示 *** */
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
  /* 确保文字不会盖住按钮，留出空间 */
  flex: 1; /* 占据剩余空间 */
  min-width: 0; /* 防止 flex item 溢出 */
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

/* *** 新增：按钮容器样式，参考 MediaLibrary *** */
.song-actions {
  display: flex;
  flex-shrink: 0;
  gap: 0.5rem;

  /* 默认状态：半透明并向右偏移，实现滑入效果 */
  opacity: 0;
  transform: translateX(10px);
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

/* *** 新增：Play Next 按钮样式，从 MediaLibrary 复制并调整 *** */
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

/* --- 删除按钮样式调整 --- */
.remove-btn {
  background: transparent;
  border: none;
  cursor: pointer;
  padding: 0; /* 调整为和 play-next-btn 一致 */
  width: 28px; /* 调整为和 play-next-btn 一致 */
  height: 28px; /* 调整为和 play-next-btn 一致 */
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #b3b3b3;
  transition: all 0.2s ease;
  /* 移除独立的 opacity 和 transform，由父级 .song-actions 控制 */
}
.remove-btn:hover {
  background-color: rgba(255, 255, 255, 0.1);
  color: #ff5555;
}
.remove-btn:active {
  transform: scale(0.95);
}

/* 剩余样式保持不变 */
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
