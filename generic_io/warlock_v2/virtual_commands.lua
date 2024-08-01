function reset_brain(ctx)
  local state, payload = ctx.commands.exec_sync('deactivate_brain')
  if state ~= 'succeeded' then
    ctx.error(payload)
  end

  local state, payload = ctx.commands.exec_sync('activate_brain')
  if state ~= 'succeeded' then
    ctx.error(payload)
  end
end
