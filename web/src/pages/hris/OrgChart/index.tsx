import { useQuery } from '@tanstack/react-query';
import { ChevronDown, ChevronRight, Users } from 'lucide-react';
import { useState } from 'react';
import { getOrgTree, type OrgNode } from '@/lib/hris/org';

export default function OrgChart() {
  const { data: tree, isLoading } = useQuery({
    queryKey: ['hris', 'org-tree'],
    queryFn: getOrgTree,
  });

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Organization Chart</h1>

      {isLoading ? (
        <div className="flex justify-center p-12">
          <span className="loading loading-spinner loading-lg" />
        </div>
      ) : tree?.length === 0 ? (
        <p className="text-center text-base-content/50 py-12">
          No departments yet. Add departments to see the org chart.
        </p>
      ) : (
        <div className="space-y-1">
          {tree?.map((node) => (
            <TreeNode key={node.id} node={node} depth={0} />
          ))}
        </div>
      )}
    </div>
  );
}

function TreeNode({ node, depth }: { node: OrgNode; depth: number }) {
  const [open, setOpen] = useState(depth < 2);
  const hasChildren = node.children && node.children.length > 0;

  return (
    <div style={{ marginLeft: depth * 24 }}>
      <button
        type="button"
        className="flex items-center gap-2 py-2 px-3 rounded-lg hover:bg-base-200 cursor-pointer"
        onClick={() => hasChildren && setOpen(!open)}
      >
        {hasChildren ? (
          open ? (
            <ChevronDown className="h-4 w-4 text-base-content/40" />
          ) : (
            <ChevronRight className="h-4 w-4 text-base-content/40" />
          )
        ) : (
          <span className="w-4" />
        )}
        <span className="font-medium">{node.name}</span>
        <span className="badge badge-sm badge-outline">{node.type}</span>
        {node.employeeCount != null && node.employeeCount > 0 && (
          <span className="text-xs text-base-content/50 flex items-center gap-1">
            <Users className="h-3 w-3" />
            {node.employeeCount}
          </span>
        )}
      </button>
      {open && hasChildren && (
        <div>
          {(node.children ?? []).map((child) => (
            <TreeNode key={child.id} node={child} depth={depth + 1} />
          ))}
        </div>
      )}
    </div>
  );
}
