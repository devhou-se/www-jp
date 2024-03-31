const reactionToEmoji = {
    '+1': '👍',
    '-1': '👎',
    'laugh': '😄',
    'confused': '😕',
    'heart': '❤️',
    'hooray': '🎉',
    'rocket': '🚀',
    'eyes': '👀'
};

// assume PostID is globally declared somewhere else on the pages
const postIDNumber = Number(PostID);
if (Number.isInteger(postIDNumber)) {
    generateReactionsElement(postIDNumber)
}

function generateReactionsElement(postId) {
    const parentElement = document.getElementById('reactions');
    const badge = document.createElement('a');
    badge.className = 'reaction-badge';
    badge.href = `https://github.com/devhou-se/www-jp/issues/${postId}`;

    parentElement.appendChild(badge)

    fetch(`https://api.devhou.se/${postId}/reactions`, {
        headers: {
            'Accept': 'application/vnd.github.squirrel-girl-preview'
        }
    })
        .then(response => response.json())
        .then(data => {
            let reactionsCount = {};
            data.forEach(reaction => {
                if (!reactionsCount[reaction.content]) {
                    reactionsCount[reaction.content] = 0;
                }
                reactionsCount[reaction.content]++;
            });

            // 反応がなかった場合は、ユーザーにこのことを知らせるメッセージを表示するだけです。
            if (Object.keys(reactionsCount).length === 0) {
                badge.textContent = 'この記事にはまだリアクションがありません';
                return;
            }

            let reactionRow = document.createElement('tr');
            for (let reaction in reactionsCount) {
                var reactionCell = document.createElement('td');
                reactionCell.textContent = `${reactionToEmoji[reaction]} ${reactionsCount[reaction]}`;
                reactionRow.appendChild(reactionCell);
            }
            badge.appendChild(reactionRow);
        })
        .catch(error => {
            console.error(error)
            badge.textContent = 'ポストリアクションを読み込めませんでした';
        });
}
