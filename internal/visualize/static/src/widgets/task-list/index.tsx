// task-list — renders tasks as a list with checkbox, id, title, deps, and wave.
//
// Deps are rendered as cross-spec anchor links (`#/spec/<dep>?tab=tasks#task-<id>`)
// so users can navigate the spec graph from the detail dialog. Each <li> also
// gets `id="task-<id>"` so cross-spec links can focus the target task after
// navigation.

import type { TaskProgress } from '@entities/spec/types';

export interface TaskListProps {
  tasks: ReadonlyArray<TaskProgress>;
}

export function TaskList({ tasks }: TaskListProps): JSX.Element {
  if (tasks.length === 0) {
    return <p>No tasks.</p>;
  }
  return (
    <ul role="list" className="task-list">
      {tasks.map((task) => (
        <li key={task.task_id} id={`task-${task.task_id}`} role="listitem">
          <input
            type="checkbox"
            checked={task.done}
            disabled
            aria-label={`task ${task.task_id} ${task.done ? 'done' : 'todo'}`}
          />
          <code>#{task.task_id}</code>
          <span className="task-title">{task.title}</span>
          {task.deps.length > 0 && (
            <span className="task-deps">
              {task.deps.map((dep) => (
                <a
                  key={dep}
                  href={`#/spec/${encodeURIComponent(dep)}?tab=tasks#task-${dep}`}
                  aria-label={`depends on #${dep} (open in this dialog)`}
                >
                  #{dep}
                </a>
              ))}
            </span>
          )}
          <span className="task-wave">wave {task.wave}</span>
        </li>
      ))}
    </ul>
  );
}