function get_hit_rate_10003(context, attack_entity, target_entity, user)
  return   10000
end

function get_fix_damage_10003(context, init_damage, attack_entity, current_entity, target_entity, user)
  init_damage =   this.get_attr_value(attack_entity,   10007)
  if damage_hit(context, 10003, attack_entity, target_entity, user) then 
    local zl001 =   math.ceil( ( init_damage * this.get_skill_damage_rate(context) ) / 10000 )
    this.add_combat_damage(context,   zl001, 10001)
  
  else
    local zl001 =   0
    this.add_combat_damage(context,   zl001, 10001)
  
  end
  return   zl001
end

