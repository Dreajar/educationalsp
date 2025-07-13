A Language Server built to learn **WHAT** LSP is and **HOW** it works.

In Neovim, hover your cursor over a word and press `Shift` + `k`

Note that kickstart.nvim comes with an LspAttach, so you don't have to write your own on_attach function. Instead you only have to write something like this:

```lua
local client = nil

vim.api.nvim_create_autocmd('FileType', {
  pattern = 'markdown',
  callback = function()
    if not client then
      local new_client = vim.lsp.start {
        name = 'educationalsp',
        cmd = { 'C:/Users/.../main.exe' },
      }
      if not new_client then
        vim.notify "Hey, you didn't do the client thing good."
        return
      end
      client = new_client
    end
    vim.lsp.buf_attach_client(0, client)
    vim.notify 'LSP attached!'
  end,
})
```
