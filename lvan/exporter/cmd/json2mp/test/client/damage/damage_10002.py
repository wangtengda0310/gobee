import math
get_hit_rate_10002(context, attack_entity, target_entity, user):
  return   max( 0, ( 10000 - this.get_attr_value(target_entity,   10065) ) )
  
get_fix_damage_10002(context, init_damage, attack_entity, current_entity, target_entity, user):
  init_damage =   this.get_attr_value(attack_entity,   10007)
  if damage_hit(context, 10002, attack_entity, target_entity, user) is True :
    # 技能基础伤害
    BaseFlow001 =   max( 1, ( math.floor( ( init_damage * context.getCustomValue('par001') ) / 10000 ) + context.getCustomValue('par002') ) )
    this.add_combat_damage(context,   BaseFlow001, 10001)
  
  else:
    BaseFlow001 =   0
    this.add_combat_damage(context,   BaseFlow001, 10001)
  return   BaseFlow001
