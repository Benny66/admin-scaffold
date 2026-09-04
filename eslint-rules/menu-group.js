// 自定义 ESLint 规则：把「菜单必须归组」编译成 AST 层红灯。
//
// 作用域：仅 frontend/src/router/index.js（由 eslint.config.js 通过 files 限定）。
// 规则检查 path === '/' 的 children 是否符合 menu-grouping 两级结构约束：
//   1. 叶子禁止裸挂顶层（除非 meta.standalone === true，为「首页/工作台」预留）
//   2. 分组节点必须有非空 meta.title 与 meta.icon
//   3. 分组的 children 不得为空
//   4. 叶子必须有 meta.title、meta.icon、meta.permission
//   5. 分组必须可达：存在 path: '' 的子项，或分组自身声明了 redirect
//
// 解析不到根路由对象时必须报错而非静默放行——延续 guard「护栏要能感知自己瞎了」铁律。

function getProperty(node, name) {
  if (!node || !node.properties) return null
  for (const prop of node.properties) {
    if (prop.type === 'Property' && prop.key) {
      const keyName =
        prop.key.type === 'Identifier'
          ? prop.key.name
          : prop.key.type === 'Literal'
            ? String(prop.key.value)
            : null
      if (keyName === name) return prop
    }
  }
  return null
}

function getPropertyValue(node, name) {
  const prop = getProperty(node, name)
  return prop ? prop.value : null
}

// 取对象字面量上某个字段的字面值（字符串/布尔/数字）；复杂表达式返回 undefined。
function getLiteralValue(node, name) {
  const val = getPropertyValue(node, name)
  if (!val) return undefined
  if (val.type === 'Literal') return val.value
  return undefined
}

// 取 meta 上某个字段的字面值
function getMetaValue(node, name) {
  const meta = getPropertyValue(node, 'meta')
  if (!meta || meta.type !== 'ObjectExpression') return undefined
  return getLiteralValue(meta, name)
}

function isString(value) {
  return typeof value === 'string'
}

function report(context, node, message) {
  context.report({ node, message })
}

const menuGroupRule = {
  meta: {
    type: 'problem',
    docs: {
      description: '菜单必须呈「分组容器 + 叶子页面」两级结构（menu-grouping）',
    },
    schema: [],
  },
  create(context) {
    return {
      // 顶层 AST：找到 const routes = [ ... ]，进而找 path: '/' 的根路由对象
      Program(programNode) {
        // 兜底：仅作用于 frontend/src/router/index.js。
        // 由 eslint.config.js 的 files/ignores 做主过滤，但若 cwd 让路径形态漂移
        // （从子目录跑 npm run lint 时路径不带 frontend/ 前缀），这里再做一次精确匹配。
        // mobile/miniapp 的 router 是单页结构（无 children），不该被本规则检查。
        const filename =
          (typeof context.filename === 'function' ? context.filename() : null) ||
          (typeof context.getFilename === 'function' ? context.getFilename() : null) ||
          ''
        if (filename && !filename.replace(/\\/g, '/').match(/\/frontend\/src\/router\/index\.js$/)) {
          return
        }
        const body = programNode.body

        // 找出 routes 数组（顶层 const routes = [...] 或 export const routes = [...]）
        let routesArray = null
        for (const stmt of body) {
          let decls = null
          if (stmt.type === 'VariableDeclaration') decls = stmt.declarations
          else if (
            stmt.type === 'ExportNamedDeclaration' &&
            stmt.declaration &&
            stmt.declaration.type === 'VariableDeclaration'
          ) {
            decls = stmt.declaration.declarations
          }
          if (!decls) continue
          for (const decl of decls) {
            if (!decl.init || decl.init.type !== 'ArrayExpression') continue
            const name =
              decl.id && decl.id.type === 'Identifier' ? decl.id.name : null
            if (name === 'routes') {
              routesArray = decl.init
              break
            }
          }
          if (routesArray) break
        }

        if (!routesArray) {
          // 文件结构变更——护栏「感知自己瞎了」
          report(
            context,
            programNode,
            'menu-grouping：未在 frontend/src/router/index.js 找到 `routes` 数组声明——路由结构写法可能已变更，请同步更新 eslint-rules/menu-group.js',
          )
          return
        }

        // 找 path === '/' 的根路由对象
        let rootRoute = null
        for (const el of routesArray.elements) {
          if (!el || el.type !== 'ObjectExpression') continue
          const pathVal = getLiteralValue(el, 'path')
          if (pathVal === '/') {
            rootRoute = el
            break
          }
        }

        if (!rootRoute) {
          report(
            context,
            programNode,
            'menu-grouping：未找到 path: \'/\' 的根路由对象——路由结构写法可能已变更，请同步更新 eslint-rules/menu-group.js',
          )
          return
        }

        const childrenProp = getProperty(rootRoute, 'children')
        if (!childrenProp || !childrenProp.value || childrenProp.value.type !== 'ArrayExpression') {
          report(
            context,
            rootRoute,
            'menu-grouping：根路由 path: \'/\' 缺少 children 数组——菜单结构必须为「分组容器 + 叶子页面」两级（见 menu-grouping design D1）',
          )
          return
        }

        const children = childrenProp.value.elements.filter(
          (n) => n && n.type === 'ObjectExpression',
        )

        children.forEach((child) => {
          const childChildrenProp = getProperty(child, 'children')
          const hasChildren =
            childChildrenProp &&
            childChildrenProp.value &&
            childChildrenProp.value.type === 'ArrayExpression'
          const isStandalone = getMetaValue(child, 'standalone') === true

          if (!hasChildren) {
            // 叶子——裸挂顶层
            if (!isStandalone) {
              report(
                context,
                child,
                'menu-grouping：菜单项禁止裸挂顶层——请归入分组（在 children 内层加 { path: \'<分组>\', meta: {...}, children: [...] }，或用 meta.standalone: true 豁免「首页/工作台」类节点）',
              )
            }
            // standalone 豁免节点跳过其余检查
            return
          }

          // 分组节点检查
          const groupChildren = childChildrenProp.value.elements.filter(
            (n) => n && n.type === 'ObjectExpression',
          )

          // 2. 分组必有非空 meta.title 与 meta.icon
          const title = getMetaValue(child, 'title')
          const icon = getMetaValue(child, 'icon')
          if (!isString(title) || title.length === 0) {
            report(
              context,
              child,
              'menu-grouping：分组节点必须声明非空 meta.title（如 meta: { title: \'系统管理\' }）',
            )
          }
          if (!isString(icon) || icon.length === 0) {
            report(
              context,
              child,
              'menu-grouping：分组节点必须声明 meta.icon（如 meta: { icon: \'Setting\' }）',
            )
          }

          // 3. 分组 children 不得为空
          if (groupChildren.length === 0) {
            report(
              context,
              childChildrenProp.value,
              'menu-grouping：分组 children 不得为空数组——请补充叶子或移除空分组',
            )
            return
          }

          // 5. 分组可达：存在 path: '' 的子项 或 分组自身声明了 redirect
          const groupRedirect = getLiteralValue(child, 'redirect')
          const hasEmptyPathChild = groupChildren.some((leaf) => {
            const p = getLiteralValue(leaf, 'path')
            return p === ''
          })
          if (!hasEmptyPathChild && !isString(groupRedirect)) {
            report(
              context,
              child,
              'menu-grouping：分组必须可达——首叶子 path: \'\'（如 { path: \'\', name: \'Asset\', component: () => import(...), meta: {...} }）或分组自身 redirect；否则直接访问分组 URL 会 404',
            )
          }

          // 4. 叶子必有 meta.title、meta.icon、meta.permission
          groupChildren.forEach((leaf) => {
            const lTitle = getMetaValue(leaf, 'title')
            const lIcon = getMetaValue(leaf, 'icon')
            const lPerm = getMetaValue(leaf, 'permission')
            if (!isString(lTitle) || lTitle.length === 0) {
              report(
                context,
                leaf,
                'menu-grouping：叶子节点必须声明非空 meta.title',
              )
            }
            if (!isString(lIcon) || lIcon.length === 0) {
              report(
                context,
                leaf,
                'menu-grouping：叶子节点必须声明 meta.icon',
              )
            }
            if (!isString(lPerm) || lPerm.length === 0) {
              report(
                context,
                leaf,
                'menu-grouping：叶子节点必须声明 meta.permission（如 meta: { permission: \'users:view\' }）——菜单按权限码过滤，缺码视为公共页除外仍要占位以表达意图',
              )
            }
          })
        })
      },
    }
  },
}

export default menuGroupRule
