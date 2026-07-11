import { RefObject, useLayoutEffect, useRef } from 'react';
import { createRoot, Root } from 'react-dom/client';
import type { Employee } from '../../api/client';
import { employeeDecorationKey } from './schedulerAvatarUtils';
import { EmployeeAvatar } from '../employees/EmployeeAvatar';

const RESOURCE_LABEL_SELECTOR = '.MuiEventCalendar-resourcesTreeItemLabel';

export function useSchedulerResourceAvatars(
  containerRef: RefObject<HTMLElement | null>,
  employees: Employee[],
) {
  const rootsRef = useRef<Map<Element, Root>>(new Map());

  useLayoutEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    const employeeByName = new Map(
      employees.map((employee) => [
        `${employee.first_name} ${employee.last_name}`.trim(),
        employee,
      ]),
    );

    const cleanupRoot = (labelEl: Element) => {
      const root = rootsRef.current.get(labelEl);
      if (root) {
        root.unmount();
        rootsRef.current.delete(labelEl);
      }
    };

    const decorateLabels = () => {
      const labels = container.querySelectorAll(RESOURCE_LABEL_SELECTOR);

      labels.forEach((labelEl) => {
        const name = labelEl.textContent?.trim() ?? '';
        const employee = employeeByName.get(name);

        if (!employee) {
          if (labelEl.getAttribute('data-meetoria-decorated')) {
            cleanupRoot(labelEl);
            labelEl.removeAttribute('data-meetoria-decorated');
          }
          return;
        }

        const decorationKey = employeeDecorationKey(employee);

        if (labelEl.getAttribute('data-meetoria-decorated') === decorationKey) {
          return;
        }

        cleanupRoot(labelEl);

        const host = labelEl as HTMLElement;
        host.textContent = '';
        host.style.display = 'inline-flex';
        host.style.alignItems = 'center';
        host.style.gap = '8px';

        const avatarMount = document.createElement('span');
        avatarMount.style.display = 'inline-flex';
        avatarMount.style.flexShrink = '0';

        const nameEl = document.createElement('span');
        nameEl.textContent = name;

        host.appendChild(avatarMount);
        host.appendChild(nameEl);
        host.setAttribute('data-meetoria-decorated', decorationKey);

        const root = createRoot(avatarMount);
        rootsRef.current.set(labelEl, root);
        root.render(
          <EmployeeAvatar
            firstName={employee.first_name}
            lastName={employee.last_name}
            avatarUrl={employee.avatar_url}
            color={employee.color}
            cacheKey={employee.updated_at}
            size={24}
          />,
        );
      });
    };

    decorateLabels();

    const observer = new MutationObserver(decorateLabels);
    observer.observe(container, { childList: true, subtree: true });

    return () => {
      observer.disconnect();
      rootsRef.current.forEach((root) => root.unmount());
      rootsRef.current.clear();
    };
  }, [containerRef, employees]);
}
