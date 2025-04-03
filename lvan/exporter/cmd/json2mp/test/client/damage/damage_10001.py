import math
get_hit_rate_10001(context, attack_entity, target_entity, user):
  return   10000
  
get_fix_damage_10001(context, init_damage, attack_entity, target_entity, user):
  init_damage =   0
  if damage_hit(context, 10001, attack_entity, target_entity, user) is True :
    this.add_combat_damage(context,   100, 10001)
  
  else:
    pass
  return   0
