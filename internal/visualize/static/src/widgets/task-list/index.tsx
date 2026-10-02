// task-list — renders tasks as a list with checkbox, id, title, deps, and wave.
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
        <li key={task.task_id} role="listitem">
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
              <code aria-label={`depends on #${task.deps.join(', #')}`}>
                {task.deps.map((d) => `#${d}`).join(', ')}
              </code>
            </span>
          )}
          <span className="task-wave">wave {task.wave}</span>
        </li>
      ))}
    </ul>
  );
}
