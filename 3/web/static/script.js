const baseUrl = '';

function loadComments(parent = null) {
let url = '/comments';
if (parent !== null) {
url += `?parent=${parent}`;
}
fetch(url)
.then(res => res.json())
.then(data => {
if (parent === null) {
document.getElementById('comments-container').innerHTML = '';
data.forEach(c => renderComment(c, document.getElementById('comments-container')));
}
});
}

function renderComment(comment, container) {
const div = document.createElement('div');
div.className = 'comment';
div.dataset.id = comment.id;

let html = `<strong>${comment.author || 'Аноним'}</strong>: ${comment.content}<br>`;
html += `<button onclick="showReplyBox(${comment.id}, this)">Ответить</button>`;
html += `<button onclick="deleteComment(${comment.id})">Удалить</button>`;
div.innerHTML = html;

const childrenDiv = document.createElement('div');
childrenDiv.className = 'children';
div.appendChild(childrenDiv);
container.appendChild(div);

if (comment.children) {
comment.children.forEach(child => renderComment(child, childrenDiv));
}
}

function addComment(parentId) {
const content = parentId ? document.getElementById(`reply-content-${parentId}`).value :
document.getElementById('new-content').value;
const author = parentId ? document.getElementById(`reply-author-${parentId}`).value :
document.getElementById('new-author').value;

if (!content.trim()) return alert('Введите текст комментария');

fetch('/comments', {
method: 'POST',
headers: {'Content-Type': 'application/json'},
body: JSON.stringify({parent_id: parentId, content, author})
})
.then(res => res.json())
.then(() => loadComments());
}

function deleteComment(id) {
if (!confirm('Удалить этот комментарий и все ответы?')) return;
fetch(`/comments/${id}`, { method: 'DELETE' })
.then(res => res.json())
.then(() => loadComments());
}

function showReplyBox(id, btn) {
if (document.getElementById(`reply-box-${id}`)) return;
const box = document.createElement('div');
box.id = `reply-box-${id}`;
box.innerHTML = `
        <textarea id="reply-content-${id}" placeholder="Ваш ответ"></textarea><br>
        <input type="text" id="reply-author-${id}" placeholder="Ваше имя">
        <button onclick="addComment(${id})">Отправить</button>
    `;
btn.parentNode.appendChild(box);
}

function searchComments() {
const term = document.getElementById('search-input').value;
fetch(`/comments?search=${encodeURIComponent(term)}`)
.then(res => res.json())
.then(data => {
document.getElementById('comments-container').innerHTML = '';
data.forEach(c => renderComment(c, document.getElementById('comments-container')));
});
}

// загружаем корневые комментарии при старте
loadComments();
